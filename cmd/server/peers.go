package main


func (s *Server) recieveFromPeer() {
	go s.pubSub.Sub.Run()
	for msg := range s.pubSub.Sub.Channel() {
		s.eventChan <- msg
	}
}

func (s *Server) registerUserKV(username string) error {
	return s.kv.Set(username, s.Config.FullAddr)
}

// peer == another server, understand ?
func (s *Server) findPeerByUsername(username string) string {
    return s.kv.Get(username)
}
