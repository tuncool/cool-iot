package conn

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego/logs"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	models2 "tunpx/lib/database/models"
	"tunpx/lib/goroutine"

	"github.com/xtaci/kcp-go"
	"tunpx/lib/common"
	"tunpx/lib/crypt"
	"tunpx/lib/pmux"
	"tunpx/lib/rate"
)

type Conn struct {
	Conn net.Conn
	Rb   []byte
}

// new conn
func NewConn(conn net.Conn) *Conn {
	return &Conn{Conn: conn}
}

func (s *Conn) readRequest(buf []byte) (n int, err error) {
	var rd int
	for {
		rd, err = s.Read(buf[n:])
		if err != nil {
			return
		}
		n += rd
		if n < 4 {
			continue
		}
		if string(buf[n-4:n]) == "\r\n\r\n" {
			return
		}
		// buf is full, can't contain the request
		if n == cap(buf) {
			err = io.ErrUnexpectedEOF
			return
		}
	}
}

// get host 、connection type、method...from connection
func (s *Conn) GetHost() (method, address string, rb []byte, err error, r *http.Request) {
	var b [32 * 1024]byte
	var n int
	if n, err = s.readRequest(b[:]); err != nil {
		return
	}
	rb = b[:n]
	r, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(rb)))
	if err != nil {
		return
	}
	hostPortURL, err := url.Parse(r.Host)
	if err != nil {
		address = r.Host
		err = nil
		return
	}
	if hostPortURL.Opaque == "443" {
		if strings.Index(r.Host, ":") == -1 {
			address = r.Host + ":443"
		} else {
			address = r.Host
		}
	} else {
		if strings.Index(r.Host, ":") == -1 {
			address = r.Host + ":80"
		} else {
			address = r.Host
		}
	}
	return
}

func (s *Conn) GetShortLenContent() (b []byte, err error) {
	var l int
	if l, err = s.GetLen(); err != nil {
		return
	}
	if l < 0 || l > 32<<10 {
		err = errors.New("read length error")
		return
	}
	return s.GetShortContent(l)
}

func (s *Conn) GetShortContent(l int) (b []byte, err error) {
	buf := make([]byte, l)
	return buf, binary.Read(s, binary.LittleEndian, &buf)
}

// 读取指定长度内容
func (s *Conn) ReadLen(cLen int, buf []byte) (int, error) {
	if cLen > len(buf) || cLen <= 0 {
		return 0, errors.New("长度错误" + strconv.Itoa(cLen))
	}
	if n, err := io.ReadFull(s, buf[:cLen]); err != nil || n != cLen {
		return n, errors.New("Error reading specified length " + err.Error())
	}
	return cLen, nil
}

func (s *Conn) GetLen() (int, error) {
	var l int32
	err := binary.Read(s, binary.LittleEndian, &l)
	return int(l), err
}

func (s *Conn) WriteLenContent(buf []byte) (err error) {
	var b []byte
	if b, err = GetLengthAndBytes(buf); err != nil {
		return
	}
	return binary.Write(s.Conn, binary.LittleEndian, b)
}

// read flag
func (s *Conn) ReadFlag() (string, error) {
	buf := make([]byte, 4)
	return string(buf), binary.Read(s, binary.LittleEndian, &buf)
}

// set alive and begin to clog
func (s *Conn) SetAlive(tp string) {
	switch s.Conn.(type) {
	case *kcp.UDPSession:
		s.Conn.(*kcp.UDPSession).SetReadDeadline(time.Time{})
	case *net.TCPConn:
		conn := s.Conn.(*net.TCPConn)
		conn.SetReadDeadline(time.Time{})
		// conn.SetKeepAlive(false)
		// conn.SetKeepAlivePeriod(time.Duration(2 * time.Second))
	case *pmux.PortConn:
		s.Conn.(*pmux.PortConn).SetReadDeadline(time.Time{})
	}
}

// set read deadline
func (s *Conn) SetReadDeadlineBySecond(t time.Duration) {
	switch s.Conn.(type) {
	case *kcp.UDPSession:
		s.Conn.(*kcp.UDPSession).SetReadDeadline(time.Now().Add(t * time.Second))
	case *net.TCPConn:
		s.Conn.(*net.TCPConn).SetReadDeadline(time.Now().Add(t * time.Second))
	case *pmux.PortConn:
		s.Conn.(*pmux.PortConn).SetReadDeadline(time.Now().Add(t * time.Second))
	}
}

// get link info from conn
func (s *Conn) GetLinkInfo() (lk *Link, err error) {
	err = s.getInfo(&lk)
	return
}

// send info for link
func (s *Conn) SendHealthInfo(info, status string) (int, error) {
	raw := bytes.NewBuffer([]byte{})
	common.BinaryWrite(raw, info, status)
	return s.Write(raw.Bytes())
}

// get health info from conn
func (s *Conn) GetHealthInfo() (info string, status bool, err error) {
	var l int
	buf := common.BufPoolMax.Get().([]byte)
	defer common.PutBufPoolMax(buf)
	if l, err = s.GetLen(); err != nil {
		return
	} else if _, err = s.ReadLen(l, buf); err != nil {
		return
	} else {
		arr := strings.Split(string(buf[:l]), common.CONN_DATA_SEQ)
		if len(arr) >= 2 {
			return arr[0], common.GetBoolByStr(arr[1]), nil
		}
	}
	return "", false, errors.New("receive health info error")
}

// TODO:获取配置信息？还是新建空信息
// get task info
func (s *Conn) GetConfigInfo() (c *models2.Client, err error) {
	err = s.getInfo(&c)
	c.Valid = true
	c.Connected = true

	// c.NoDisplay = false
	return
}

// get task info
// get task info
func (s *Conn) GetTunnelInfo() (t *models2.Tunnel, err error) {
	// t = new(file.Tunnel)
	err = s.getInfo(&t)
	// t.Id = database.GetDb().JsonDb.GetTaskId()
	// t.NoStore = true
	t.Flow = new(models2.Flow)
	t.Target = new(models2.Target)
	// TODO:这里判断白名单端口
	// t.Target.TargetStr = beego.AppConfig.String("allow_ports")
	// if len(t.Target.TargetStr) < 2 {
	//	t.Target.TargetStr = "81-65535"
	// }
	return
}

// send  info
func (s *Conn) SendInfo(t interface{}, flag string) (int, error) {
	/*
		The task info is formed as follows:
		+----+-----+---------+
		|type| len | content |
		+----+---------------+
		| 4  |  4  |   ...   |
		+----+---------------+
	*/
	raw := bytes.NewBuffer([]byte{})
	if flag != "" {
		binary.Write(raw, binary.LittleEndian, []byte(flag))
	}
	b, err := json.Marshal(t)
	if err != nil {
		return 0, err
	}
	lenBytes, err := GetLengthAndBytes(b)
	if err != nil {
		return 0, err
	}
	binary.Write(raw, binary.LittleEndian, lenBytes)
	return s.Write(raw.Bytes())
}

// get task info
func (s *Conn) getInfo(t interface{}) (err error) {
	var l int
	buf := common.BufPoolMax.Get().([]byte)
	defer common.PutBufPoolMax(buf)
	if l, err = s.GetLen(); err != nil {
		return
	} else if _, err = s.ReadLen(l, buf); err != nil {
		return
	} else {
		// 客户端向服务端发送的信息，使用的是json格式
		json.Unmarshal(buf[:l], &t)
	}
	return
}

// close
func (s *Conn) Close() error {
	return s.Conn.Close()
}

// write
func (s *Conn) Write(b []byte) (int, error) {
	return s.Conn.Write(b)
}

// read
func (s *Conn) Read(b []byte) (n int, err error) {
	if s.Rb != nil {
		// if the rb is not nil ,read rb first
		if len(s.Rb) > 0 {
			n = copy(b, s.Rb)
			s.Rb = s.Rb[n:]
			return
		}
		s.Rb = nil
	}
	return s.Conn.Read(b)
}

// write sign flag
func (s *Conn) WriteClose() (int, error) {
	return s.Write([]byte(common.RES_CLOSE))
}

// write main
func (s *Conn) WriteMain() (int, error) {
	return s.Write([]byte(common.WORK_MAIN))
}

// write main
func (s *Conn) WriteConfig() (int, error) {
	return s.Write([]byte(common.WORK_CONFIG))
}

// write chan
func (s *Conn) WriteChan() (int, error) {
	return s.Write([]byte(common.WORK_CHAN))
}

// get task or host result of add
func (s *Conn) GetAddStatus() (b bool) {
	binary.Read(s.Conn, binary.LittleEndian, &b)
	return
}

func (s *Conn) WriteAddOk() error {
	return binary.Write(s.Conn, binary.LittleEndian, true)
}

// 添加隧道失败
func (s *Conn) WriteAddFail() error {
	defer s.Close()
	return binary.Write(s.Conn, binary.LittleEndian, false)
}

func (s *Conn) LocalAddr() net.Addr {
	return s.Conn.LocalAddr()
}

func (s *Conn) RemoteAddr() net.Addr {
	return s.Conn.RemoteAddr()
}

func (s *Conn) SetDeadline(t time.Time) error {
	return s.Conn.SetDeadline(t)
}

func (s *Conn) SetWriteDeadline(t time.Time) error {
	return s.Conn.SetWriteDeadline(t)
}

func (s *Conn) SetReadDeadline(t time.Time) error {
	return s.Conn.SetReadDeadline(t)
}

// get the assembled amount data(len 4 and content)
func GetLengthAndBytes(buf []byte) (b []byte, err error) {
	raw := bytes.NewBuffer([]byte{})
	if err = binary.Write(raw, binary.LittleEndian, int32(len(buf))); err != nil {
		return
	}
	if err = binary.Write(raw, binary.LittleEndian, buf); err != nil {
		return
	}
	b = raw.Bytes()
	return
}

// udp connection setting
func SetUdpSession(sess *kcp.UDPSession) {
	sess.SetStreamMode(true)
	sess.SetWindowSize(1024, 1024)
	sess.SetReadBuffer(64 * 1024)
	sess.SetWriteBuffer(64 * 1024)
	sess.SetNoDelay(1, 10, 2, 1)
	sess.SetMtu(1600)
	sess.SetACKNoDelay(true)
	sess.SetWriteDelay(false)
}

// conn1 mux conn
func CopyWaitGroup(conn1, conn2 net.Conn, encrypt bool, snappy bool, rate *rate.Rate, isServer bool, rb []byte, task *models2.Tunnel) {
	// var in, out int64
	// var wg sync.WaitGroup
	connHandle := GetConn(conn1, encrypt, snappy, rate, isServer)
	if rb != nil {
		connHandle.Write(rb)
	}
	flow := &models2.Flow{ClientId: task.ClientId}
	wg := new(sync.WaitGroup)
	wg.Add(1)
	err := goroutine.CopyConnsPool.Invoke(goroutine.NewConns(connHandle, conn2, flow, wg, task))
	wg.Wait()
	if err != nil {
		logs.Error(err)
	}
}

// get crypt or snappy conn
func GetConn(conn net.Conn, cpt, snappy bool, rt *rate.Rate, isServer bool) io.ReadWriteCloser {
	if cpt {
		if isServer {
			return rate.NewRateConn(crypt.NewTlsServerConn(conn), rt)
		}
		return rate.NewRateConn(crypt.NewTlsClientConn(conn), rt)
	} else if snappy {
		return rate.NewRateConn(NewSnappyConn(conn), rt)
	}
	return rate.NewRateConn(conn, rt)
}

type LenConn struct {
	conn io.Writer
	Len  int
}

func NewLenConn(conn io.Writer) *LenConn {
	return &LenConn{conn: conn}
}

func (c *LenConn) Write(p []byte) (n int, err error) {
	n, err = c.conn.Write(p)
	c.Len += n
	return
}
