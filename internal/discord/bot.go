package discord

type Bouncer interface {
	check(ID string) bool
}

type GuildBouncer struct {
	guildIDs map[string]bool
}

func NewGuildBouncer(guildIDs []string) *GuildBouncer {
	guildIDMap := make(map[string]bool)
	for _, id := range guildIDs {
		guildIDMap[id] = true
	}
	return &GuildBouncer{guildIDs: guildIDMap}
}

func (b *GuildBouncer) check(ID string) bool {
	if _, ok := b.guildIDs[ID]; ok {
		return true
	}
	return false
}

type ChannelBouncer struct {
	channelIDs map[string]bool
}

func NewChannelBouncer(channelIDs []string) *ChannelBouncer {
	channelIDMap := make(map[string]bool)
	for _, id := range channelIDs {
		channelIDMap[id] = true
	}
	return &ChannelBouncer{channelIDs: channelIDMap}
}

func (b *ChannelBouncer) check(ID string) bool {
	if _, ok := b.channelIDs[ID]; ok {
		return true
	}
	return false
}

type Bot struct {
	guildBouncer   Bouncer
	channelBouncer Bouncer
	session        *Session
}

func NewBot(guildBouncer Bouncer, channelBouncer Bouncer, session *Session) (*Bot, error) {
	return &Bot{
		guildBouncer:   guildBouncer,
		channelBouncer: channelBouncer,
		session:        session,
	}, nil
}

func (b *Bot) Shutdown() {
	if b.session != nil {
		b.session.Close()
		b.session = nil
	}
}

func (b *Bot) IsGuildAllowed(guildID string) bool {
	if b.guildBouncer != nil {
		return b.guildBouncer.check(guildID)
	}
	return true
}

func (b *Bot) IsChannelAllowed(channelID string) bool {
	if b.channelBouncer != nil {
		return b.channelBouncer.check(channelID)
	}
	return true
}
