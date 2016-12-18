package mqtt

type SessionManager struct {
	Sessions map[string]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		Sessions: make(map[string]*Session),
	}
}

func (this *SessionManager) New(id string) (*Session, error) {
	return &Session{}, nil
}

func (this *SessionManager) Add(id string, sess *Session) {
	this.Sessions[id] = sess
}

func (this *SessionManager) Get(id string) (*Session, error) {
	return this.Sessions[id], nil
}
