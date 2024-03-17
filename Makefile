seed:
	rm db.sqlite
	sqlite3 db.sqlite < setup.sql
