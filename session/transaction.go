package session

import "geeorm/log"

// Begin 开始一个数据库事务
//
// Begin 方法用于开始一个新的数据库事务。
// 它会记录事务开始的日志，并调用底层数据库的 Begin 方法。
// 如果事务开始失败，会记录错误日志并返回错误。
//
// 返回值:
//   - err: 如果事务开始失败，返回错误信息。
func (s *Session) Begin() (err error) {
	log.Info("transaction begin")
	if s.tx, err = s.db.Begin(); err != nil {
		log.Error(err)
		return
	}
	return
}

// Commit 提交当前事务
//
// Commit 方法用于提交当前的数据库事务。
// 它会记录事务提交的日志，并调用底层事务的 Commit 方法。
// 如果提交失败，会记录错误日志并返回错误。
//
// 返回值:
//   - err: 如果提交失败，返回错误信息。
func (s *Session) Commit() (err error) {
	log.Info("transaction commit")
	if err = s.tx.Commit(); err != nil {
		log.Error(err)
	}
	return
}

// Rollback 回滚当前事务
//
// Rollback 方法用于回滚当前的数据库事务。
// 它会记录事务回滚的日志，并调用底层事务的 Rollback 方法。
// 如果回滚失败，会记录错误日志并返回错误。
//
// 返回值:
//   - err: 如果回滚失败，返回错误信息。
func (s *Session) Rollback() (err error) {
	log.Info("transaction rollback")
	if err = s.tx.Rollback(); err != nil {
		log.Error(err)
	}
	return
}
