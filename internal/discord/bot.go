package discord

const (
	WATCHLIST_ADD    = "watchlist_add"
	WATCHLIST_REMOVE = "watchlist_remove"
	WATCHLIST_VIEW   = "watchlist_view"
	WATCHLIST_CLEAR  = "watchlist_clear"
)

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

func (b *GuildBouncer) Check(ID string) bool {
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

func (b *ChannelBouncer) Check(ID string) bool {
	if _, ok := b.channelIDs[ID]; ok {
		return true
	}
	return false
}
