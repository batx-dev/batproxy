package sql

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/batx-dev/batproxy"
)

const defaultPageSize = 1000

type ProxyService struct {
	db *DB
}

func NewProxy(db *DB) *ProxyService {
	return &ProxyService{db: db}
}

var _ batproxy.ProxyService = (*ProxyService)(nil)

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
	if opts.UUID != "" {
		where, args = append(where, "uuid = ?"), append(args, opts.UUID)
	}

	var pageToken int
	if len(opts.PageToken) > 0 {
		if pageToken, err = strconv.Atoi(opts.PageToken); err != nil {
			return nil, fmt.Errorf("PageToken is invalid")
		}
	}
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT 
		    uuid,
		    user,
		    host,
		    private_key,
		    passphrase,
		    password,
		    node,
		    port,
		    create_time,
		    update_time
		FROM t_bat_proxy WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(pageSize, pageToken),
		args...,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	proxies := make([]*batproxy.Proxy, 0)
	for rows.Next() {
		proxy := &batproxy.Proxy{}
		if err = rows.Scan(
			&proxy.UUID,
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
			return nil, fmt.Errorf("sacn t_bat_proxy: %v", err)
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
