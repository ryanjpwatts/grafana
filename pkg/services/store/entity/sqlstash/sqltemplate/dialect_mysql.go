package sqltemplate

// MySQL is the default implementation of Dialect for the MySQL DMBS, currently
// supporting MySQL-8.x. It relies on having ANSI_QUOTES SQL Mode enabled. For
// more information about ANSI_QUOTES and SQL Modes see:
//
//	https://dev.mysql.com/doc/refman/8.4/en/sql-mode.html#sqlmode_ansi_quotes
var MySQL = mysql{
	rowLockingClauseMap: rowLockingClauseAll,
	argPlaceholderFunc:  argFmtSQL92,
	name:                "mysql8",
}

// MySQL5 is the implementation of Dialect for the MySQL DMBS providing support
// for the latest 5.x version, currently 5.7.
var MySQL5 = mysql{
	rowLockingClauseMap: rowLockingClauseMap{
		SelectForShare:            SelectForShare,
		SelectForShareNoWait:      SelectForShare,
		SelectForShareSkipLocked:  SelectForShare,
		SelectForUpdate:           SelectForUpdate,
		SelectForUpdateNoWait:     SelectForUpdate,
		SelectForUpdateSkipLocked: SelectForUpdate,
	},
	argPlaceholderFunc: argFmtSQL92,
	name:               "mysql5",
}

var _ Dialect = MySQL

type mysql struct {
	standardIdent
	rowLockingClauseMap
	argPlaceholderFunc
	name
}
