package commands

import (
	"fmt"
	"math/rand"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"

	"whatsmeow-bot/utils"
)

func init() {
    RegisterCommand(&PackgodCommand{})
}

type PackgodCommand struct{}

func (c *PackgodCommand) Execute(client *whatsmeow.Client, message *events.Message, args []string) *string {
    var nouns = []string{
        "toaster",
        "garden gnome",
        "wet sock",
        "expired coupon",
        "mop handle",
        "busted printer",
        "crusty sponge",
        "flip phone",
        "traffic cone",
        "deflated balloon",
        "off-brand cereal",
        "fax machine",
        "footrest",
        "unplugged tv",
        "soggy burrito",
        "broken usb stick",
        "warped cutting board",
        "dusty VHS tape",
        "melted candle",
        "backwards hat rack",
        "fridge magnet",
        "squeaky shopping cart",
        "dented kettle",
        "plastic spork",
        "rickety ladder",
        "outlet cover",
        "sticky mousepad",
        "blunt pencil",
        "empty tissue box",
        "mismatched sock",
        "paperclip chain",
        "chewed pen cap",
        "ragged oven mitt",
        "dead flashlight",
        "clogged drain",
        "unicorn band-aid",
        "wrinkled receipt",
        "cracked plate",
        "sad piñata",
        "festival porta-potty",
        "foam mannequin head",
        "tangled extension cord",
        "broken zipper",
        "spilled glitter",
        "loose floorboard",
        "wobbly table leg",
        "flat bike tire",
        "off-tune kazoo",
        "mop handle",
        "busted printer",
        "crusty sponge",
        "flip phone",
        "traffic cone",
        "deflated balloon",
        "off-brand cereal",
        "fax machine",
        "footrest",
        "unplugged tv",
        "soggy burrito",
        "broken usb stick",
        "warped cutting board",
        "fridge magnet",
        "toenail clipping",
        "plastic fork",
        "sticky note",
        "couch crumb",
        "loading screen",
        "burnt toast",
        "flat soda",
        "blunt pencil",
        "lego instruction manual",
        "half-used tissue",
        "tv static",
        "pebble in a shoe",
        "alarm clock on monday",
        "dent in a can",
        "squeaky door hinge",
        "stale popcorn",
        "elevator music",
        "empty lollipop wrapper",
        "out-of-ink pen",
        "yogurt lid",
        "shopping cart wheel",
        "crumb-filled keyboard",
        "dryer lint",
        "dead pixel",
        "ice cube in a urinal",
        "traffic light on red",
        "wrinkled napkin",
        "missed call notification",
        "used band-aid",
        "sock without a pair",
        "warm milk",
        "public restroom door handle",
        "chewed pencil eraser",
        "rusty spoon",
        "refrigerator smell",
        "bootleg dvd",
        "microwave beep",
        "broken umbrella",
        "cable knot",
        "splintered chopstick",
        "watered-down soup",
        "glitched-out npc",
        "cracked phone screen",
        "empty stapler",
        "random sock in the laundry",
        "dull butterknife",
        "fidget spinner",
        "wall scuff",
        "soggy toast",
        "half-loaded dishwasher",
        "wobbly chair leg",
        "hood ornament",
        "manual can opener",
        "slow wifi signal",
        "cheap halloween mask",
        "itty-bitty shampoo bottle",
        "receipt from 2014",
        "dust bunny",
        "worn-out welcome mat",
        "blinking smoke alarm",
        "creaky floorboard",
        "off-brand charger",
        "crusty pillowcase",
        "puddle in the driveway",
        "tangled necklace",
        "shoelace",
        "ticket stub",
        "remote battery",
        "bathroom fan",
        "zipper that always jams",
        "toothpick in a steak",
    }
    var responses = []string{
        "you built like a %v", 
        "you look like a %v", 
        "you have the personality of a %v",
        "your mum fucked a %v and had you",

        "you built like a %v had a baby with a %v", 
        "you look like a %v mixed with a %v",
    }
    
    responseIndex := rand.Intn(len(responses))
    response := responses[responseIndex]
    noun1, noun2 := nouns[rand.Intn(len(nouns))], nouns[rand.Intn(len(nouns))]

    var text string
    if responseIndex < 4 {
        text = fmt.Sprintf(response, noun1)
    } else {
        text = fmt.Sprintf(response, noun1, noun2)
    }

    resp := utils.Reply(client, message, text)
	if resp != nil {
		utils.React(client, message, "✅")
	} else {
		utils.React(client, message, "❌")
	}

    return nil
}

func (c *PackgodCommand) Name() string {
    return "packgod"
}

func (c *PackgodCommand) Description() string {
    return "Replies with a nonsense 'pack'"
}