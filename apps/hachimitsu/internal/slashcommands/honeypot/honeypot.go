// Package honeypot contains the hachimitsu /honeypot slash command group.
package honeypot

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hachimitsu/internal/ent"
	"jurien.dev/yugen/hachimitsu/internal/services"
	"jurien.dev/yugen/hachimitsu/internal/slashcommands/honeypot/add"
	"jurien.dev/yugen/hachimitsu/internal/slashcommands/honeypot/edit"
	"jurien.dev/yugen/hachimitsu/internal/slashcommands/honeypot/list"
	"jurien.dev/yugen/hachimitsu/internal/slashcommands/honeypot/remove"
	localStatic "jurien.dev/yugen/hachimitsu/internal/static"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/utils"
)

var errChannelConflict = errors.New("channel already a honeypot")

type honeypotSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
}

// modalInputs holds the parsed component values from a honeypot modal submit.
type modalInputs struct {
	channelID string
	days      int
	roleIDs   []string
	errMsg    string
}

// HoneypotModule is the /honeypot command group. It also owns the modal submit
// handlers for /honeypot add and /honeypot edit.
type HoneypotModule struct {
	container  *di.Container
	honeypot   *services.HoneypotService
	subModules []honeypotSubModule
}

// GetHoneypotModule constructs the HoneypotModule and all its leaf sub-modules.
func GetHoneypotModule(container *di.Container) *HoneypotModule {
	hpSvc := container.Get(localStatic.DiHoneypot).(*services.HoneypotService)

	return &HoneypotModule{
		container: container,
		honeypot:  hpSvc,
		subModules: []honeypotSubModule{
			add.GetAddModule(container),
			edit.GetEditModule(container),
			remove.GetRemoveModule(container),
			list.GetListModule(container),
		},
	}
}

// Commands returns the top-level /honeypot command registration.
func (m *HoneypotModule) Commands() []disgoplus.CommandRegistration {
	opts := make([]discord.ApplicationCommandOption, 0, len(m.subModules))
	for _, sub := range m.subModules {
		opts = append(opts, sub.SubCommandOption())
	}

	return []disgoplus.CommandRegistration{
		disgoplus.Global(discord.SlashCommandCreate{
			Name:        "honeypot",
			Description: "Manage honeypot channels",
			Options:     opts,
		}),
	}
}

// Register wires all sub-command handlers and the modal submit handlers.
func (m *HoneypotModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildAdminMiddleware)

		for _, sub := range m.subModules {
			sub.Register(r)
		}
	})

	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildAdminMiddleware)
		r.Modal("/HONEYPOT_ADD", m.handleAddModal)
		r.Modal("/HONEYPOT_EDIT/{id}", m.handleEditModal)
	})
}

// parseHoneypotModal extracts channel, days, and roles from a modal submit.
// When errMsg is non-empty the parse failed and errMsg should be shown to the
// user.
func parseHoneypotModal(
	data discord.ModalSubmitInteractionData,
) modalInputs {
	channelSel, ok := data.ChannelSelectMenu("channel")
	if !ok || len(channelSel.Values) == 0 {
		return modalInputs{errMsg: "Please select a channel."}
	}

	days, parseErr := strconv.Atoi(data.Text("days"))
	if parseErr != nil || days < 0 || days > 7 {
		return modalInputs{
			errMsg: "Days must be a number between 0 and 7.",
		}
	}

	roleSel, _ := data.RoleSelectMenu("roles")

	roleIDs := make([]string, len(roleSel.Values))
	for i, id := range roleSel.Values {
		roleIDs[i] = id.String()
	}

	return modalInputs{
		channelID: channelSel.Values[0].String(),
		days:      days,
		roleIDs:   roleIDs,
	}
}

// handleAddModal processes /honeypot add modal submissions.
func (m *HoneypotModule) handleAddModal(e *handler.ModalEvent) error {
	if e.GuildID() == nil {
		return nil
	}

	ctx := context.Background()
	guildID := e.GuildID().String()

	inputs := parseHoneypotModal(e.Data)
	if inputs.errMsg != "" {
		if cerr := e.CreateMessage(discord.MessageCreate{
			Content: inputs.errMsg,
			Flags:   discord.MessageFlagEphemeral,
		}); cerr != nil {
			return fmt.Errorf("honeypot add modal: create message: %w", cerr)
		}

		return nil
	}

	alreadyExists, createErr := m.createNewHoneypot(
		ctx, guildID, inputs.channelID, inputs.days, inputs.roleIDs,
	)
	if createErr != nil {
		if cerr := e.CreateMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); cerr != nil {
			return fmt.Errorf("honeypot add modal: create message: %w", cerr)
		}

		return nil
	}

	if alreadyExists {
		if cerr := e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"<#%s> is already a honeypot. "+
					"Use `/honeypot edit` to change its settings.",
				inputs.channelID,
			),
			Flags: discord.MessageFlagEphemeral,
		}); cerr != nil {
			return fmt.Errorf("honeypot add modal: create message: %w", cerr)
		}

		return nil
	}

	daysLabel := "days"
	if inputs.days == 1 {
		daysLabel = "day"
	}

	if cerr := e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"🍯 <#%s> is now a honeypot! "+
				"Anyone posting there without an exempt role will be "+
				"banned and have %d %s of messages deleted.",
			inputs.channelID,
			inputs.days,
			daysLabel,
		),
		Flags: discord.MessageFlagEphemeral,
	}); cerr != nil {
		return fmt.Errorf("honeypot add modal: create message: %w", cerr)
	}

	return nil
}

// createNewHoneypot checks for an existing entry, creates a new one if absent,
// and applies the ignored role list. Returns alreadyExists=true when the
// channel is already registered.
func (m *HoneypotModule) createNewHoneypot(
	ctx context.Context,
	guildID, channelID string,
	days int,
	roleIDs []string,
) (alreadyExists bool, err error) {
	existing, err := m.honeypot.Get(ctx, guildID, channelID)
	if err != nil {
		return false, fmt.Errorf("get: %w", err)
	}

	if existing != nil {
		return true, nil
	}

	hp, createErr := m.honeypot.Create(ctx, guildID, channelID, days)
	if createErr != nil {
		return false, fmt.Errorf("create: %w", createErr)
	}

	if len(roleIDs) > 0 {
		if setErr := m.honeypot.SetIgnoredRoles(
			ctx, hp.ID, roleIDs,
		); setErr != nil {
			utils.Logger.Warnf(
				"honeypot add modal: set roles %s/%s: %v",
				guildID,
				channelID,
				setErr,
			)
		}
	}

	return false, nil
}

// handleEditModal processes /honeypot edit modal submissions.
func (m *HoneypotModule) handleEditModal(e *handler.ModalEvent) error {
	if e.GuildID() == nil {
		return nil
	}

	ctx := context.Background()
	guildID := e.GuildID().String()

	recordID, parseErr := strconv.Atoi(e.Vars["id"])
	if parseErr != nil {
		if cerr := e.CreateMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); cerr != nil {
			return fmt.Errorf(
				"honeypot edit modal: create message: %w", cerr,
			)
		}

		return nil
	}

	hp, err := m.honeypot.GetByID(ctx, recordID)
	if err != nil || hp == nil {
		if cerr := e.CreateMessage(discord.MessageCreate{
			Content: "Honeypot not found. It may have been removed.",
			Flags:   discord.MessageFlagEphemeral,
		}); cerr != nil {
			return fmt.Errorf(
				"honeypot edit modal: create message: %w", cerr,
			)
		}

		return nil
	}

	inputs := parseHoneypotModal(e.Data)
	if inputs.errMsg != "" {
		if cerr := e.CreateMessage(discord.MessageCreate{
			Content: inputs.errMsg,
			Flags:   discord.MessageFlagEphemeral,
		}); cerr != nil {
			return fmt.Errorf(
				"honeypot edit modal: create message: %w", cerr,
			)
		}

		return nil
	}

	if err := m.applyHoneypotEdit(
		ctx, guildID, hp, inputs,
	); err != nil {
		msg := "Something went wrong, try again later."
		if errors.Is(err, errChannelConflict) {
			msg = fmt.Sprintf(
				"<#%s> is already a honeypot channel.",
				inputs.channelID,
			)
		}

		if cerr := e.CreateMessage(discord.MessageCreate{
			Content: msg,
			Flags:   discord.MessageFlagEphemeral,
		}); cerr != nil {
			return fmt.Errorf(
				"honeypot edit modal: create message: %w", cerr,
			)
		}

		return nil
	}

	daysLabel := "days"
	if inputs.days == 1 {
		daysLabel = "day"
	}

	if cerr := e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"🍯 <#%s> updated — banning with **%d %s** of messages deleted.",
			inputs.channelID,
			inputs.days,
			daysLabel,
		),
		Flags: discord.MessageFlagEphemeral,
	}); cerr != nil {
		return fmt.Errorf("honeypot edit modal: create message: %w", cerr)
	}

	return nil
}

// applyHoneypotEdit validates the channel change and persists all edits.
// Returns errChannelConflict when the new channel is already a honeypot.
func (m *HoneypotModule) applyHoneypotEdit(
	ctx context.Context,
	guildID string,
	hp *ent.Honeypot,
	inputs modalInputs,
) error {
	if inputs.channelID != hp.ChannelID {
		conflict, conflErr := m.honeypot.Get(
			ctx, guildID, inputs.channelID,
		)
		if conflErr != nil {
			return fmt.Errorf("get conflict check: %w", conflErr)
		}

		if conflict != nil {
			return errChannelConflict
		}
	}

	if _, updateErr := m.honeypot.Update(
		ctx,
		hp.ID,
		func(u *ent.HoneypotUpdateOne) {
			u.SetChannelID(inputs.channelID).
				SetDeleteMessageDays(inputs.days).
				SetIgnoredRoleIDs(inputs.roleIDs)
		},
	); updateErr != nil {
		return fmt.Errorf("update: %w", updateErr)
	}

	return nil
}
