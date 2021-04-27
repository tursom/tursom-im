package im_conn

type void struct{}

var member void

type ConnGroup struct {
	connList map[*AttachmentConn]void
}

func NewConnGroup() *ConnGroup {
	return &ConnGroup{
		connList: make(map[*AttachmentConn]void),
	}
}

func (g ConnGroup) Add(conn *AttachmentConn) {
	g.connList[conn] = member
}

func (g ConnGroup) Remove(conn *AttachmentConn) {
	delete(g.connList, conn)
}

func Write(bytes byte[]) {

}
