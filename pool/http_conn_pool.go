package http_server

/* tcp连接池 */
func init() {
	for i := 0; i < poolSize; i++ {

		conn, err := net.Dial("tcp", "localhost:50000")
		if err != nil {
			fmt.Println(err)
			fmt.Println("init conn pool failed!", err)
		}
		pool <- conn
	}

}

func borrow() (conn net.Conn, err error) {
	select {
	case conn := <-pool:
		return conn, err
	}
	return
}

func release(conn net.Conn) error {
	select {
	case pool <- conn:
		return nil
	}

}

