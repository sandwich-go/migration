### v0.1.0-alpha.9 (2024-07-23 15:11:02)

#### 🌎 Downloads
  * Docker : 
	* **CenturyGame**: harbor.centurygame.com/zhongtai/migration:0.1.0-alpha.9

#### 🛠  Refactor
  * add log ([616b26d](https://bitbucket.org/funplus/migration/commits/616b26d22d86196f69c8b3fd70d8e7cc14ef9c87)) (<small>[Daming Yang](daming.yang@centurygame.com)@2024-07-23 15:11:02 &#43;0800 &#43;0800</small>)

### v0.1.0-alpha.8 (2024-03-04 13:12:25)

#### 🌎 Downloads
  * Docker : 
	* **CenturyGame**: harbor.centurygame.com/zhongtai/migration:0.1.0-alpha.8

#### 🚀  New Feature
  * 增加获取 hints 接口 ([22cf717](https://bitbucket.org/funplus/migration/commits/22cf717c6e87bec7057c343a30b80cc566f47350)) (<small>[huangqing.zhu](huangqing.zhu@centurygame.com)@2024-03-04 13:12:25 &#43;0800 &#43;0800</small>)

### v0.1.0-alpha.7 (2023-11-20 12:33:04)

#### 🌎 Downloads
  * Docker : 
	* **CenturyGame**: harbor.centurygame.com/zhongtai/migration:0.1.0-alpha.7

#### 🐛  Bug Fixed
  * 当migration执行报错，输出日志 ([6b8020c](https://bitbucket.org/funplus/migration/commits/6b8020c6f7f46efed05ab4278a7bb39fbf10ca3d)) (<small>[huangqing.zhu](huangqing.zhu@centurygame.com)@2023-11-20 12:33:04 &#43;0800 &#43;0800</small>)

### v0.1.0-alpha.6 (2023-10-25 18:30:01)

#### 🌎 Downloads
  * Docker : 
	* **CenturyGame**: harbor.centurygame.com/zhongtai/migration:0.1.0-alpha.6

#### 🛠  Refactor
  * show ddl 移除文件写入 ([1482bfc](https://bitbucket.org/funplus/migration/commits/1482bfcfa26d17bfb0b266f192cda26724c5b913)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2023-10-25 18:30:01 &#43;0800 &#43;0800</small>)

### v0.1.0-alpha.5 (2023-10-18 10:48:46)

#### 🌎 Downloads
  * Docker : 
	* **CenturyGame**: harbor.centurygame.com/zhongtai/migration:0.1.0-alpha.5

#### 🐛  Bug Fixed
  * migration message ([163e043](https://bitbucket.org/funplus/migration/commits/163e0432c314a2b182ee46bbd89732b0c1178ce0)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2023-10-18 10:48:46 &#43;0800 &#43;0800</small>)

#### 🤖  Tools
  * **sem**: make changelog ([d242261](https://bitbucket.org/funplus/migration/commits/d2422610c72ee8b2f5078bcc91ca8b57f4999441)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-09-13 19:02:12 &#43;0800 &#43;0800</small>)

### v0.1.0-alpha.4 (2023-09-13 19:01:10)

#### 🌎 Downloads
  * Docker : 
	* **CenturyGame**: harbor.centurygame.com/zhongtai/migration:0.1.0-alpha.4

#### 🐛  Bug Fixed
  * showDDL add filter ([17edd66](https://bitbucket.org/funplus/migration/commits/17edd663c397646da012cbd194269bbc69f5abd2)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-09-13 19:01:10 &#43;0800 &#43;0800</small>)

#### 🤖  Tools
  * **sem**: make changelog ([fdaaae4](https://bitbucket.org/funplus/migration/commits/fdaaae4e426b04a61ffae215c53ce2dc84608d61)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-09-11 15:35:33 &#43;0800 &#43;0800</small>)

### v0.1.0-alpha.3 (2023-09-11 15:34:42)

#### 🌎 Downloads
  * Docker : 
	* **CenturyGame**: harbor.centurygame.com/zhongtai/migration:0.1.0-alpha.3

#### 🐛  Bug Fixed
  * 增加隐藏日志中数据库密码功能 ([da58bdd](https://bitbucket.org/funplus/migration/commits/da58bdd33ddf0277f081989acffbdcedd8366b0f)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-09-11 15:34:42 &#43;0800 &#43;0800</small>)

#### 🤖  Tools
  * **sem**: make changelog ([f295d02](https://bitbucket.org/funplus/migration/commits/f295d021505c37f61c901357e7430642c710ea03)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-09-11 12:59:02 &#43;0800 &#43;0800</small>)

### v0.1.0-alpha.2 (2023-09-11 12:52:30)

#### 🌎 Downloads
  * Docker : 
	* **CenturyGame**: harbor.centurygame.com/zhongtai/migration:0.1.0-alpha.2

#### 🐛  Bug Fixed
  * update .sembumprc.yml ([6f6e0dd](https://bitbucket.org/funplus/migration/commits/6f6e0ddafa7fe7f076d0c35dffd885f97b949c4b)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-09-11 12:52:30 &#43;0800 &#43;0800</small>)
  * 所有接口通过prepare()进入到migration目录后，最后需要调用prepare()返回的deferFunc()函数返回起始目录 ([9f89051](https://bitbucket.org/funplus/migration/commits/9f890514a989099d1d1a2a7b65cb5b1f159b268d)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-09-11 12:33:32 &#43;0800 &#43;0800</small>)

#### 🤖  Tools
  * **sem**: make changelog ([935bee0](https://bitbucket.org/funplus/migration/commits/935bee00594c4053945c2fc71fb42a2d7ef9c7db)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-05-19 21:41:03 &#43;0800 &#43;0800</small>)

#### 💪  Commit
  * Merge branch 'feature/migration' into version/0.1 ([e19a5ff](https://bitbucket.org/funplus/migration/commits/e19a5ff3984a9f24df50718ae273ffb240313812)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-09-11 12:33:47 &#43;0800 &#43;0800</small>)

### v0.1.0-alpha.1 (2023-05-19 21:40:37)

#### 🌎 Downloads
  * Docker : 
	* **CenturyGame**: harbor.centurygame.com/zhongtai/migration:0.1.0-alpha.1

#### 🐛  Bug Fixed
  * 当migrations/versions目录不存在时 需要手动创建 ([60b5cca](https://bitbucket.org/funplus/migration/commits/60b5cca095cfe44fe5b8a353b90b8850bd518958)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-05-19 21:40:05 &#43;0800 &#43;0800</small>)
  * 掉版本插入语句 ([4734ba6](https://bitbucket.org/funplus/migration/commits/4734ba66f7ba60e79f05d8d1bc1381ea6e20b3aa)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-05-19 17:59:13 &#43;0800 &#43;0800</small>)
  * 优化日志格式 ([49812c4](https://bitbucket.org/funplus/migration/commits/49812c418fab2e7c5c8da9d70e8122412ee55536)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-05-18 14:40:30 &#43;0800 &#43;0800</small>)
  * 去掉migration工作目录/CommitID/层 ([2c30520](https://bitbucket.org/funplus/migration/commits/2c305202fb363037cbd852143c482880ebba5efd)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-05-18 14:39:57 &#43;0800 &#43;0800</small>)
  * flask db migrate 增加rev-id参数指定版本号为commitId;message内容格式为:commitId_ts ([516f032](https://bitbucket.org/funplus/migration/commits/516f032b5feb299ad9d00f728d5efcdf924cb028)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-05-16 20:03:16 &#43;0800 &#43;0800</small>)

#### 💪  Commit
  * Merge branch 'feature/migration' into version/0.1 ([451592b](https://bitbucket.org/funplus/migration/commits/451592b23e38231ee0cdcec4e8cecf7243b227e5)) (<small>[zhagnxujun](xujun.zhang@centurygame.com)@2023-05-19 21:40:37 &#43;0800 &#43;0800</small>)



