package bdkeeper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	// _ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	// _ "github.com/jackc/pgx/v5/stdlib"
	"github.com/wurt83ow/gophermart/internal/models"
	"github.com/wurt83ow/gophermart/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log interface {
	Info(string, ...zapcore.Field)
}

type BDKeeper struct {
	conn *sql.DB
	log  Log
}

func NewBDKeeper(dsn func() string, log Log) *BDKeeper {
	addr := dsn()
	if addr == "" {
		log.Info("database dsn is empty")

		return nil
	}

	conn, err := sql.Open("pgx", dsn())
	if err != nil {
		log.Info("Unable to connection to database: ", zap.Error(err))

		return nil
	}

	driver, err := postgres.WithInstance(conn, new(postgres.Config))
	if err != nil {
		log.Info("error getting driver: ", zap.Error(err))

		return nil
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Info("error getting getwd: ", zap.Error(err))
	}

	// fix error test path
	mp := dir + "/migrations"

	var path string
	if _, err := os.Stat(mp); err != nil {
		path = "../../"
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%smigrations", path),
		"postgres",
		driver)
	if err != nil {
		log.Info("Error creating migration instance : ", zap.Error(err))
	}

	err = m.Up()
	if err != nil {
		log.Info("Error while performing migration: ", zap.Error(err))
	}

	log.Info("Connected!")

	return &BDKeeper{
		conn: conn,
		log:  log,
	}
}

func (kp *BDKeeper) GetUserWithdrawals(userID string) ([]models.DataWithdraw, error) {
	ctx := context.Background()

	// get withdrawals from bd
	sql := `
	SELECT
		id_order_out AS order,
		- sum(accrual) AS sum,
		processed_at AS data 
	FROM
		savings_account
	WHERE
		accrual < 0
		AND user_id = $1
	GROUP BY
		id_order_out,
		processed_at
	ORDER BY
		processed_at`

	rows, err := kp.conn.QueryContext(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user withdrawls by userID: %w", err)
	}

	defer rows.Close()

	result := make([]models.DataWithdraw, 0)

	for rows.Next() {
		var m models.DataWithdraw

		err := rows.Scan(&m.Order, &m.Sum, &m.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to get user withdrawls by userID: %w", err)
		}

		m.DateRFC = m.Date.Format(time.RFC3339)
		result = append(result, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get user withdrawls by userID: %w", err)
	}

	return result, nil
}

func (kp *BDKeeper) GetUserBalance(userID string) (models.DataBalance, error) {
	ctx := context.Background()

	sql := `
	SELECT
		SUM(sq.current) AS current,
		- SUM(sq.withdrawn) AS withdrawn
	FROM (
		SELECT
			SUM(accrual)
			CURRENT,
			0 withdrawn
		FROM
			savings_account
		WHERE
			user_id = $1
		UNION
		SELECT
			0,
			SUM(accrual)
		FROM
			savings_account
		WHERE
			user_id = $1
			AND accrual < 0) AS sq`
	row := kp.conn.QueryRowContext(ctx, sql, userID)

	// read the values from the database record into the corresponding fields of the structure
	var m models.DataBalance

	err := row.Scan(&m.Current, &m.Withdrawn)
	if err != nil {
		kp.log.Info("row scan error: ", zap.Error(err))

		return models.DataBalance{}, fmt.Errorf("failed to get user balance by userID: %w", err)
	}

	return m, nil
}

func (kp *BDKeeper) GetOpenOrders() ([]string, error) {
	ctx := context.Background()

	// get orders from bd
	sql := `
	SELECT
		number
	FROM
		public.orders
	WHERE
		status <> 'INVALID'
		AND status <> 'PROCESSED'
		AND number <> ''
	LIMIT 100`

	rows, err := kp.conn.QueryContext(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get open orders: %w", err)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get open orders: %w", err)
	}

	defer rows.Close()

	orders := make([]string, 0)

	for rows.Next() {
		var m models.ExtRespOrder

		err := rows.Scan(&m.Order)
		if err != nil {
			return nil, fmt.Errorf("failed to get open orders: %w", err)
		}

		orders = append(orders, m.Order)
	}

	return orders, nil
}

func (kp *BDKeeper) LoadOrders() (storage.StorageOrders, error) {
	ctx := context.Background()

	// get orders from bd
	sql := `
	SELECT
		o.order_id,
		o.number,
		o.status,
		o.date,		 
		COALESCE(s.accrual, 0) AS accrual,
		o.user_id
	FROM
		orders AS o
		LEFT JOIN savings_account AS s ON o.order_id = s.id_order_in
			AND o.date = s.processed_at`

	rows, err := kp.conn.QueryContext(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to load orders: %w", err)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to load orders: %w", err)
	}

	defer rows.Close()

	data := make(storage.StorageOrders)

	for rows.Next() {
		var m models.DataOrder

		err := rows.Scan(&m.UUID, &m.Number,
			&m.Status, &m.Date, &m.Accrual, &m.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to load orders: %w", err)
		}

		m.DateRFC = m.Date.Format(time.RFC3339)
		data[m.Number] = m
	}

	return data, nil
}

func (kp *BDKeeper) LoadUsers() (storage.StorageUsers, error) {
	ctx := context.Background()

	// get users from bd
	sql := `
	SELECT
		user_id,
		name,
		email,
		hash
	FROM
		users`

	rows, err := kp.conn.QueryContext(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to load users: %w", err)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to load users: %w", err)
	}

	defer rows.Close()

	data := make(storage.StorageUsers)

	for rows.Next() {
		var m models.DataUser

		err := rows.Scan(&m.UUID, &m.Name, &m.Email, &m.Hash)
		if err != nil {
			return nil, fmt.Errorf("failed to load users: %w", err)
		}

		data[m.Email] = m
	}

	return data, nil
}

func (kp *BDKeeper) SaveOrder(key string, order models.DataOrder) (models.DataOrder, error) {
	ctx := context.Background()

	var id string

	if order.UUID == "" {
		neuuid := uuid.New()
		id = neuuid.String()
	} else {
		id = order.UUID
	}

	sql := `
	INSERT INTO orders (order_id, number, date, status, user_id)
		VALUES ($1, $2, $3, $4, $5)
	RETURNING
		order_id`
	_, err := kp.conn.ExecContext(ctx, sql,
		id, order.Number, order.Date, order.Status, order.UserID)

	sql = `
	SELECT
		d.order_id,
		d.number,
		d.date,
		d.status,
		d.user_id
	FROM
		orders d
	WHERE
		d.number = $1`
	row := kp.conn.QueryRowContext(ctx, sql, order.Number)

	// read the values from the database record into the corresponding fields of the structure
	var m models.DataOrder

	nerr := row.Scan(&m.UUID, &m.Number, &m.Date, &m.Status, &m.UserID)
	if nerr != nil {
		kp.log.Info("row scan error: ", zap.Error(err))

		return order, fmt.Errorf("failed to save order: %w", nerr)
	}

	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			kp.log.Info("unique field violation on column: ", zap.Error(err))

			return m, storage.ErrConflict
		}

		return m, fmt.Errorf("failed to save order: %w", err)
	}

	return m, nil
}

func (kp *BDKeeper) SaveUser(key string, data models.DataUser) (models.DataUser, error) {
	ctx := context.Background()

	var id string

	if data.UUID == "" {
		neuuid := uuid.New()
		id = neuuid.String()
	} else {
		id = data.UUID
	}

	sql := `
	INSERT INTO users (user_id, email, hash, name)
		VALUES ($1, $2, $3, $4)
	RETURNING
		user_id`
	_, err := kp.conn.ExecContext(ctx, sql,
		id, data.Email, data.Hash, data.Name)

	var (
		cond string
		hash []byte
	)

	if data.Hash != nil {
		cond = "AND u.hash = $2"
		hash = data.Hash
	}

	sql = `
	SELECT
		u.user_id,
		u.email,
		u.hash,
		u.name
	FROM
		users u
	WHERE
		u.email = $1 %s`
	sql = fmt.Sprintf(sql, cond)
	row := kp.conn.QueryRowContext(ctx, sql, data.Email, hash)

	// read the values from the database record into the corresponding fields of the structure
	var m models.DataUser

	nerr := row.Scan(&m.UUID, &m.Email, &m.Hash, &m.Name)
	if nerr != nil {
		return data, fmt.Errorf("failed to save user: %w", nerr)
	}

	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			kp.log.Info("unique field violation on column: ", zap.Error(err))

			return m, storage.ErrConflict
		}

		return m, fmt.Errorf("failed to save user: %w", err)
	}

	return m, nil
}

func (kp *BDKeeper) Withdraw(withdraw models.DataWithdraw) error {
	ctx := context.Background()

	// start the transaction
	tx, err := kp.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	// if the commit is unsuccessful, all changes to the transaction will be rolled back
	defer tx.Rollback()

	Args := []interface{}{withdraw.UserID}

	// 1.Создадим дополнительный select, чтобы заблокировать все записи
	// покупателя для изменения, так как "FOR UPDATE" не работате со
	// сгруппированными строками.
	// 2.Получим таблицу всех баллов по покупателю в разрезе заказов,
	// а так же, в колонке user_accruel все баллы по покупателю в целом.
	// 3.Соединим сгруппированную таблицу со вложенным запросом, чтобы получить
	// все накопленные баллы покупателя.
	// 4. Упорядочим строки по дате заказа (получим дату дополнительным левым соединением)
	sql := `
	WITH _orders AS (
		SELECT
			*
		FROM
			orders
		WHERE
			user_id = $1
		FOR UPDATE
	)
	SELECT
		sa.user_id,
		sa.id_order_in AS number,
		_orders.date AS date,
		SUM(sa.accrual) AS accrual,
		nq.user_accrual
	FROM
		savings_account AS sa
		INNER JOIN _orders AS _orders ON sa.id_order_in = _orders.number
		INNER JOIN (
			SELECT
				user_id,
				SUM(accrual) AS user_accrual
			FROM
				savings_account
			WHERE
				user_id = $1
			GROUP BY
				user_id) AS nq ON nq.user_id = sa.user_id
	WHERE
		sa.user_id = $1
	GROUP BY
		sa.user_id,
		sa.id_order_in,
		_orders.date,
		nq.user_accrual
	ORDER BY
		_orders.date ASC`

	rows, err := tx.QueryContext(ctx, sql, Args...)
	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	if rows.Err() != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	defer rows.Close()

	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)

	leftWrite := withdraw.Sum
	idx := 0

	for rows.Next() {
		if leftWrite <= 0 {
			break
		}

		var m models.DataOrder

		err := rows.Scan(&m.UserID, &m.Number,
			&m.Date, &m.Accrual, &m.UserAccrual)
		if err != nil {
			return fmt.Errorf("failed to withdraw: %w", err)
		}

		// Вернем ошибку, если сумма всех накопленных баллов пользователя меньше, чем
		// сумма запрошенная к списанию
		if m.UserAccrual < withdraw.Sum {
			return storage.ErrInsufficient
		}

		// Создадим строки с минусом для каждой строки заказа
		// и вычтем сумму списания из "ОсталосьСписать"
		accrual := float32(math.Min(float64(leftWrite), float64(m.Accrual)))
		leftWrite -= accrual

		valueStrings = append(valueStrings,
			fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
				idx*5+1, idx*5+2, idx*5+3, idx*5+4, idx*5+5))

		valueArgs = append(valueArgs, withdraw.UserID)
		valueArgs = append(valueArgs, time.Now())
		valueArgs = append(valueArgs, m.Number)
		valueArgs = append(valueArgs, withdraw.Order)
		valueArgs = append(valueArgs, -accrual)
		idx++
	}

	// Запишем набор на списание баллов с минусом.
	sql = `
	INSERT INTO savings_account (user_id, processed_at, id_order_in, id_order_out, accrual)
    VALUES %s`
	sql = fmt.Sprintf(sql, strings.Join(valueStrings, ","))
	_, err = kp.conn.ExecContext(ctx, sql, valueArgs...)

	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	// На всякий случай, после записи нашего набора, проверим остаток по покупателю в целом
	// и если он вдруг окажется меньше нуля, то вернем ошибку, следовательно не произойдет фиксация
	// транзакции, она откатится и запись в базу будет отменена.
	sql = `
	SELECT
		SUM(accrual) AS accrual
	FROM
		savings_account
	WHERE
		user_id = $1`
	row := kp.conn.QueryRowContext(ctx, sql, withdraw.UserID)

	// read the values from the database record into the corresponding fields of the structure
	var m models.ExtRespOrder
	err = row.Scan(&m.Accrual)

	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	if m.Accrual < 0 {
		return storage.ErrInsufficient
	}

	// commit the transaction
	return tx.Commit()
}

func (kp *BDKeeper) UpdateOrderStatus(result []models.ExtRespOrder) error {
	ctx := context.Background()

	valueStrings := make([]string, 0, len(result))
	valueArgs := make([]interface{}, 0, len(result)*2)

	for i, v := range result {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, v.Order)
		valueArgs = append(valueArgs, v.Status)
	}

	sql := `
	WITH _data (
		number,
		status
	) AS (
		VALUES % s)
	UPDATE
		orders
	SET
		status = CAST(_data.status AS statuses)
	FROM
		_data
	WHERE
		orders.number = _data.number`
	sql = fmt.Sprintf(sql, strings.Join(valueStrings, ","))

	_, err := kp.conn.ExecContext(ctx, sql, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (kp *BDKeeper) InsertAccruel(orders map[string]models.ExtRespOrder) error {
	ctx := context.Background()

	valueStrings := make([]string, 0, len(orders))
	valueArgs := make([]interface{}, 0, len(orders)*2)

	i := 0
	for _, v := range orders {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, v.Order)
		valueArgs = append(valueArgs, fmt.Sprintf("%f", v.Accrual))
		i++
	}

	sql := `
	WITH _data (
		number,
		accrual
	) AS (
		VALUES % s)
	INSERT INTO savings_account (user_id, processed_at, id_order_in, accrual)
	SELECT
		orders.user_id,
		CURRENT_TIMESTAMP,
		_data.number,
		to_number(_data.accrual, '999G9999D99999999')
	FROM
		_data
		INNER JOIN orders ON _data.number = orders.number
		LEFT JOIN savings_account AS SA ON _data.number = SA.id_order_in
			AND SA.id_order_out IS NULL
	WHERE
		SA.id_order_in IS NULL`
	sql = fmt.Sprintf(sql, strings.Join(valueStrings, ","))

	_, err := kp.conn.ExecContext(ctx, sql, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (kp *BDKeeper) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	defer cancel()

	if err := kp.conn.PingContext(ctx); err != nil {
		return false
	}

	return true
}

func (kp *BDKeeper) Close() bool {
	kp.conn.Close()

	return true
}
