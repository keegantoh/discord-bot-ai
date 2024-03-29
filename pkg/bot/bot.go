package bot

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	discord "github.com/bwmarrin/discordgo"
)

type Bot struct {
	*discord.Session

	Router *Router
}

func NewBot(token string) (*Bot, error) {
	session, err := discord.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	return &Bot{
		Session: session,
		Router:  NewRouter(nil),
	}, nil
}

func (b *Bot) Run(guildID string, removeCommands bool) {
	b.Identify.Intents = discord.MakeIntent((discord.IntentsAllWithoutPrivileged | discord.IntentMessageContent))
	b.AddHandler(func(s *discord.Session, r *discord.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	b.AddHandler(b.Router.HandleInteraction)
	b.AddHandler(b.Router.HandleMessage)

	err := b.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	err = b.Router.Sync(b.Session, guildID)
	if err != nil {
		panic(err)
	}
	defer b.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	if removeCommands {
		log.Println("Removing commands...")
		b.Router.ClearCommands(b.Session, guildID)
	}
	log.Println(("Gracefully shutting down."))
}
