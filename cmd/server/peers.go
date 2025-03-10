package main


func (s *Server) recieveFromPeer() {
	go s.pubSub.Sub.Run()
	for msg := range s.pubSub.Sub.Channel() {
        msg.FromPeer = true
		s.eventChan <- msg
	}
}

func (s *Server) registerUserKV(username string) error {
	return s.kv.Set(username, s.Config.FullAddr)
}

