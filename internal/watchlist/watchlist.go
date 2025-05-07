package watchlist

type Watchlist struct {
	songIDs []string
}

func NewWatchlist() *Watchlist {
	return &Watchlist{
		songIDs: []string{},
	}
}

func (w *Watchlist) Add(songID string) {
	w.songIDs = append(w.songIDs, songID)
}

func (w *Watchlist) Remove(songID string) {
	for i, id := range w.songIDs {
		if id == songID {
			w.songIDs = append(w.songIDs[:i], w.songIDs[i+1:]...)
			break
		}
	}
}

func (w *Watchlist) List() []string {
	return w.songIDs
}

func (w *Watchlist) Clear() {
	w.songIDs = []string{}
}
