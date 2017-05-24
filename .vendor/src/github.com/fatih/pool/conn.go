package pool

import "github.com/mavricknz/ldap"

// PoolConn is a wrapper around net.Conn to modify the the behavior of
// net.Conn's Close() method.
type PoolConn struct {
	DB       *ldap.LDAPConnection
	c        *channelPool
	unusable bool
}

// Close() puts the given connects back to the pool instead of closing it.
func (p *PoolConn) Close() error {
	if p == nil {
		return nil
	}
	if p.unusable {
		if p.DB != nil {
			if p.DB.Connected {
				return p.DB.Close()
			}
		}
		return nil
	}
	return p.c.put(p.DB)
}

// MarkUnusable() marks the connection not usable any more, to let the pool close it instead of returning it to pool.
func (p *PoolConn) MarkUnusable() {
	p.unusable = true
}

// newConn wraps a standard net.Conn to a poolConn net.Conn.
func (c *channelPool) wrapConn(conn *ldap.LDAPConnection) *PoolConn {
	p := &PoolConn{c: c}
	p.DB = conn
	return p
}
