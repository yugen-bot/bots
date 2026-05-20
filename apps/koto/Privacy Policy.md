## Privacy Policy

### Usage of Data

The bot may use stored data, as defined below, for different features including but not limited to: Welcoming joining users (when enabled) and command handling.  
No usage of data outside of the aformentioned cases will happen and the data is not shared with any 3rd-party site or service.

### Stored Information

The bot may store the following information automatically when being invited to a new Discord Server:

-   `id` with the Server's ID as value.

The id + default values for the settings are required for the bot to function.

Users partaking in any of the Koto games will have the following information stored after participation:

-   `id` with the User's ID as a value.
-   `participated` with the amount of games it participated in as value.
-   `wins` with the amount of games it won as value.
-   `inGuild` whether the user is still part of the Server.

### Updating Data

The data may be updated when using specific commands.  
Updating data will require the input of an end user, and data that can be seen as sensitive, such as content of a message, may need to be stored when using certain commands.

Participation in games can also update the User's information, this will only update the 3 mentioned keys excluding `id` stated earlier.

### Temporarely stored Information

The Bot may keep the stored information in an internal cacheing mechanic for a certain amount of time.  
After this time period, the cached information will be dropped and only be re-added when required.

Data may be dropped from cache pre-maturely through actions such as removing the bot from the Server.

### Removal of Data

#### Automatic removal

Stored Server Data can be removed automatically through means of removing the bot from a Server. This can be achieved by either kicking or banning the bot from the server. Re-inviting the bot will add the same default values, as mentioned above, back to the bot's database.

User's data can not be removed automatically.

#### Manual removal

Manual removal of the data can be requested through email at [info@jurien.dev](mailto:info@jurien.dev).
For security reasons will we ask you to provide us with proof of ownership of the server, that you wish the data to be removed of. Only a server owner may request manual removal of data and requesting it will result in the bot being removed from the server, if still present on it.
