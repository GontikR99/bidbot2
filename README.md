# BidBot2: an EverQuest/Discord utility for handling DKP auctions

BidBot2 is a Windows application written in Go and Lua, which is designed to
make running DKP-based loot auctions in EverQuest more efficient and managable.

While some attempts have been made to keep complexity down, this is still
very much a work in progress.  There are still a lot of sharp edges, and the bot is
still likely to crash at all the wrong moments.  Use at your own risk.

## Prerequisites: Accounts
BidBot2 requires dedicated accounts with three different services: EverQuest,
Discord, and Google's Cloud Text-to-Speech.

### EverQuest:
BidBot2 is designed to run with full control of an EverQuest client.  If you try
to play on a machine at the same time you're running the bot on the machine, you
might end up with some undesired effects.

### Discord:
BidBot2 requires a Discord bot token.  You can get one for free at: 

* https://discordapp.com/developers/applications/me

A walkthrough for how to set up a Discord application, create a bot, and
retrieve your bot's token, and invite the bot to your Discord guild
can be found at:

* https://github.com/Chikachi/DiscordIntegration/wiki/How-to-get-a-token-and-channel-ID-for-Discord

### Google Cloud Text-to-Speech
BidBot2 requires a paid account with Google, as it is designed to use the
Google Cloud Text-to-Speech service.  As I write this, Google is charging
US$0.16 for each megabyte of text converted to speech, and gives you the
first megabyte for free.

While this is a paid service, BidBot2 makes very little use of the service.  A
month of heavy raiding and occasional testing by the author ran up 36KiB worth 
of usage -- less than 4% of the free allowance.

You will need to obtain a file "google-account.json".  Follow steps 1-4 on the
following page:

*  https://cloud.google.com/text-to-speech/docs/quickstart-client-libraries

## Initial setup
Using BidBit2 requires some setup.  First an EverQuest character needs to be
set up.  Then BidBot2 itself needs to be configured.  Finally, if you wish
to have BidBot2 use item links during auctions, some additional steps
are required.

### In EverQuest
EverQuest must be run in windowed mode for BidBot2 to control it.

To use BidBot2, you should create an EverQuest character dedicated to it.
Create this character, invite it to your guild, and the open the options page (Alt-O).
Turn the following options OFF:

* Use Tell Windows
* Join General Channels
* Auto Show Rewards
* Auto Turn On AFK
* (Confirmations) Raid Invite
* Blink Active Chat Window

Additionally:

* Set `Allow trading with` to `No one`
* Set `Current Font` to `Ariel` if that font is not already selected.
* Click `Load UI Skin`, select `default`, and then click `Load Skin`

You also might wish to change some of the advanced video settings to improve your framerate.

If you changed the font, you should log out of EverQuest and back in before
continuing.

Once you've set up EverQuest, go find a nice wall to stare at.  BidBot2 expects a reasonably high framerate,
and might have problems if the framerate is too low.

### In BidBot2
For the first run, BidBot2 needs some information:

* `EverQuest directory`: Enter the location of your EverQuest installation.
* `Announcement Channel`: Auctions are sent to the selected channel.  Guild is default.  
For testing, you might wish to select "Say".
* `Control channel:password`: BidBot2 expects to receive privileged commands
over a password protected in-game channel.  Specify the name of the channel and
the password here.  Example: `mybidbot:thepassword`.
* `Link items during auction`: Disable this for now, we'll get back to this shortly.
* `Discord Token`: your Discord bot token, from the prerequisites.
* `Google Cloud TTS credentials`: Enter the location of your `google-account.json` 
file, from the prerequisites.
* `Rules script`: Enter the location of the Lua script which tells the bot
how to get DKP information and decide auctions.

 With your character logged in and setup as in the previous section, and with the BidBot configured as
 above, press the "Start" button, and watch the bot set up to go.
 
### In Discord:
Once BidBot2 is running, it's time to set up the connection to Discord.  If you haven't yet invited
the bot to your Discord server, there will appear a URL in the bottom portion of the BidBot2 window that
you can copy/paste into your web browser to invite the bot to your server.
 
Go to your guild's Discord.  Join the text channel you wish BidBot2 to write in, and join the
voice channel you wish BidBot2 to speak in.  Then send two messages to the channel:
 
* `!bindtext`
* `!bindvoice`
 
If BidBot2 is running and set up properly, it will respond to each message, telling you that it has
successfully bound to the specified channels.
 
## Running an auction
Log into EverQuest on a different computer, and join the same chat channel that BidBot2 has joined.
Then send to that channel the text `!auc` followed (on the same line) by an item link.  BidBot2 will
conduct an auction of the specified item, and report the results in Discord and in EverQuest.
 
## Bot commands
BidBot2 responds to three different types of commands.
 
### In Discord
 
* `!bindtext`: Set the Discord text channel that BidBot2 will report auctions to.  Only Discord server
 administrators are permitted to issue this command.
* `!bindvoice`: Set the Discord voice channel that BidBot2 will announce auctions in.  Only Discord server
 administrators are permitted to issue this command.
* `!dkp <character name>`:  Ask BidBot2 to look up the current DKP total for the specified character.
 
### Tells sent in EverQuest
BidBot2 responds to the following commands when any player sends them to BidBot2 as an EverQuest
tell message.
 
* `!dkp`: Ask BidBot2 to look up the current DKP total for the character sending the tell.
 
### Messages sent to the Command and Control channel
BidBot2 responds to the following commands when any player sends them to BidBot2's command and
control channel.  Be sure not to give the password to this channel to anyone who shouldn't issue
the corresponding command.
 
* `!auc <item link>`: Run an auction for the specified item
* `!calibrate`: Perform one-time calibration when using link-items mode (see below)
* `!echo <item link>`: Send the controller a tell with the specified item link to text item linking
in link-items mode (see below)
 
## Linking items during auctions
By default, BidBot2 will not attempt to link items back to EverQuest when running auctions.  This is
because BidBot2 requires some special setup to reliably click on item links.  Before enabling 
`Link items during auction`, check the following:
 
* Verify that EverQuest is running in windowed mode on your primary monitor.  BidBot2 can only see and
click on your primary monitor.
* Ensure the EverQuest is not being scaled up by Windows.  If your primary monitor is a very high
resolution (such as a 4k monitor), Windows will scale the EverQuest window by up to 200%.  BidBot2
cannot read EverQuest if EverQuest has been scaled up.

Assuming you pass those checks, you can perform one-time item-linking setup: 

* Log entirely out of EverQuest, and stop BidBot2.
* Enable the `Link items during auction` checkbox.
* Select the character that BidBot2 will be controlling from the `Bot character` pulldown.
* Click the `Write window layout` button.  BEWARE: This will completely obliterate the bot character's
current window layout, and replace it with a layout friendly to BidBot2
* Log back into your BidBot2 controlled character in EverQuest.
* Start BidBot2 with the `Start` button.
* On a different machine, send the command `!calibrate` to BidBot2's command and control channel.  It
may take more than one attempt for this to succeed.  If you failed to set the window layout, if
EverQuest is running in full-screen mode, or if EverQuest is being scaled by Windows, this step
will never succeed.  When it succeeds, BidBot2 will respond in tells, saying `Calibration complete.`

Once this setup is complete, you can verify it is working:

* Send the text `!echo` followed by an item link to BidBot2's command and control channel.  If
successful, BidBot2 will respond to you by tell with the same item you linked in the command.
* Assuming the `!echo` test succeeds, BidBot2 is not properly set up to link items during auctions.

## Customizing BidBot2 for your guild: Lua rules.
**To be written**
