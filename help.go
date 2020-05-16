package main

import (
	"fmt"
	"log"
)

// ZW is a zero-width space
const ZW = string(0x200b)

func genHelpText(conf *Config) string {
	helpText := fmt.Sprintf("  ~~ dungeonbot %s ~~\n", VERSION)
	helpText += `github.com/gbmor/dungeonbot

An assistance bot for tabletop RPG games being played through IRC.

    !roll NdN[+-N]
        Roll dice or a die with optional modifier. Eg: !roll 1d20+4
        The die type must be one of [d4|d6|d8|d10|d12|d20|d100]
        and number less than 100 in quantity.

    !add [campaign] $NAME
        Add a campaign notepad called $NAME

    !adduser [campaign] $NAME $NICK
        Add $NICK to the list of users authorized to make changes to
        a campaign notepad called $NAME

    !append [campaign] $NAME $NOTE
        Append $NOTE to a campaign notepad called $NAME. A blank line
        will separate each note entry.

    !campaign $NAME
        Retrieve the campaign notepad for $NAME`
	helpText += "\n"

	helpCtx := fmt.Sprintf("I'm an assistance bot for tabletop RPG games made by g%sbmor. You probably want to '!roll 1d20+4', but for extended help see: ", ZW)

	url, err := pastebin(conf.pastebinURL, helpText)
	if err != nil {
		log.Printf("When sending help text: %s", err.Error())
		url = "%error in pastebin service%"
	}

	return fmt.Sprintf("%s%s", helpCtx, url)
}
