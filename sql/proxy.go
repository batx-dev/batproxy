package sql

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/batx-dev/batproxy"
	"github.com/mattn/go-sqlite3"
	"k8s.io/apimachinery/pkg/util/rand"
)

const defaultPageSize = 1000

type ProxyService struct {
	db *DB

	// suffix used for create proxy
	suffix string
}

type ProxyServiceOptions struct {
	Suffix string
}

func NewProxy(db *DB, opts ProxyServiceOptions) *ProxyService {
	return &ProxyService{db: db, suffix: opts.Suffix}
}

var _ batproxy.ProxyService = (*ProxyService)(nil)

func (s *ProxyService) CreateProxy(ctx context.Context, proxy *batproxy.Proxy, opts batproxy.CreateProxyOptions) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if opts.Suffix == "" {
		opts.Suffix = s.suffix
	}

	if err := createProxy(ctx, tx, proxy, opts); err != nil {
		return err
	}

	return tx.Commit()
}

func createProxy(ctx context.Context, tx *Tx, proxy *batproxy.Proxy, opts batproxy.CreateProxyOptions) error {
	if err := proxy.Validate(); err != nil {
		return batproxy.Errorf(batproxy.EINVALID, "%v", err)
	}

	if i := strings.Index(proxy.Host, ":"); i == -1 {
		proxy.Host += ":22"
	}

	if strings.HasPrefix(proxy.ID, "http://") {
		proxy.ID = strings.TrimPrefix(proxy.ID, "http://")
	}
	if strings.HasPrefix(proxy.ID, "https://") {
		proxy.ID = strings.TrimPrefix(proxy.ID, "https://")
	}

	if proxy.ID == "" {
		proxy.ID = rand.String(8)

		if opts.Suffix != "" {
			if !strings.HasPrefix(opts.Suffix, ".") {
				proxy.ID += "."
			}

			proxy.ID += opts.Suffix
		}
	}

	proxy.CreateTime = tx.now
	proxy.UpdateTime = proxy.CreateTime

	_, err := tx.ExecContext(ctx, `
		INSERT INTO "t_bat_proxy" (
			proxy_id, 
		    user, 
		    host,
		    private_key, 
		    passphrase, 
		    password, 
		    node,
		    port, 
		    create_time, 
		    update_time
		)  
		VALUES (?,?,?,?,?,?,?,?,?,?)
		`,
		&proxy.ID,
		&proxy.User,
		&proxy.Host,
		&proxy.PrivateKey,
		&proxy.Passphrase,
		&proxy.Password,
		&proxy.Node,
		&proxy.Port,
		&proxy.CreateTime,
		&proxy.UpdateTime,
	)
	if err != nil {
		if drvErr, ok := err.(sqlite3.Error); ok {
			if drvErr.Code == sqlite3.ErrConstraint &&
				drvErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return batproxy.Errorf(batproxy.ECONFLICT, "'%s' already exists", proxy.ID)
			}
		}

		// TODO: support mysql
		return err
	}

	return nil
}

func (s *ProxyService) ListProxies(ctx context.Context, opts batproxy.ListProxiesOptions) (page *batproxy.ListProxiesPage, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if page, err = listProxies(ctx, tx, opts); err != nil {
		return nil, err
	}

	return page, nil
}

func listProxies(ctx context.Context, tx *Tx, opts batproxy.ListProxiesOptions) (page *batproxy.ListProxiesPage, err error) {
	var args []interface{}
	where := []string{"1 = 1"}
	if opts.ProxyID != "" {
		where, args = append(where, "proxy_id = ?"), append(args, opts.ProxyID)
	}

	var pageToken int
	if len(opts.PageToken) > 0 {
		if pageToken, err = strconv.Atoi(opts.PageToken); err != nil {
			return nil, batproxy.Errorf(batproxy.EINVALID, "page_token is invalid")
		}
	}
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT 
		    proxy_id,
		    user,
		    host,
		    private_key,
		    passphrase,
		    password,
		    node,
		    port,
		    create_time,
		    update_time
		FROM "t_bat_proxy" WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(pageSize, pageToken),
		args...,
	)

	if err != nil {
		return nil, fmt.Errorf("select 't_bat_proxy': %v", err)
	}
	defer rows.Close()

	proxies := make([]*batproxy.Proxy, 0)
	for rows.Next() {
		proxy := &batproxy.Proxy{}
		if err = rows.Scan(
			&proxy.ID,
			(*NullString)(&proxy.User),
			(*NullString)(&proxy.Host),
			(*NullString)(&proxy.PrivateKey),
			(*NullString)(&proxy.Passphrase),
			(*NullString)(&proxy.Password),
			(*NullString)(&proxy.Node),
			&proxy.Port,
			(*NullTime)(&proxy.CreateTime),
			(*NullTime)(&proxy.UpdateTime),
		); err != nil {
			return nil, fmt.Errorf("sacn 't_bat_proxy': %v", err)
		}
		proxies = append(proxies, proxy)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %v", err)
	}

	page = &batproxy.ListProxiesPage{
		Proxies: proxies,
	}

	if len(proxies) == pageSize {
		page.NextPageToken = strconv.Itoa(pageToken + pageSize)
	}

	return page, nil
}

func (s *ProxyService) DeleteProxy(ctx context.Context, proxyID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	page := &batproxy.ListProxiesPage{}
	proxy := &batproxy.Proxy{}

	if page, err = listProxies(ctx, tx, batproxy.ListProxiesOptions{ProxyID: proxyID}); err != nil {
		return err
	}

	if len(page.Proxies) != 0 {
		proxy = page.Proxies[0]
	}

	if err := deleteProxy(ctx, tx, proxyID); err != nil {
		return err
	}

	s.db.Logger.V(1).Info("delete",
		"proxy_id", proxyID,
		"user", proxy.User,
		"host", proxy.Host,
		"node", proxy.Node,
		"port", proxy.Port,
	)

	return tx.Commit()
}

func deleteProxy(ctx context.Context, tx *Tx, proxyID string) error {
	if proxyID == "" {
		return batproxy.Errorf(batproxy.EINVALID, "field proxy id is required")
	}
	_, err := tx.ExecContext(ctx, `
		DELETE FROM "t_bat_proxy"
		WHERE proxy_id = ?
	`, proxyID)
	return err
}
