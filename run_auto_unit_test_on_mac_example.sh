DSN="user:password@tcp(hostname)/database?charset=utf8mb4&parseTime=True&loc=Local"

sudo UNIT_TEST_MYSQL_DSN_GAEA=$DSN $GOPATH/bin/goconvey -port=8079
