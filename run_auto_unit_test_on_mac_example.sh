DSN="user:password@tcp(hostname)/database?charset=utf8mb4&parseTime=True&loc=Local"

# port=端口 packages=要并行测试的包数, 为了避免数据库出现脑裂, 配置为 1 excludedDirs=被排除在监视范围之外的目录, 测试通过之后的包不在重复进行测试, 加快开发时的测试速度
sudo UNIT_TEST_MYSQL_DSN_GAEA=$DSN $GOPATH/bin/goconvey -port=8079 -packages=1 -excludedDirs=config,orm,encrypt,logger,request,router,utt,verify,wechat,events,manages,services
