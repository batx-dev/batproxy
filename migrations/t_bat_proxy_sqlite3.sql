CREATE TABLE IF NOT EXISTS `t_bat_proxy` (
  `id` INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  `proxy_id` varchar(128) NOT NULL,
  `user` varchar(128) DEFAULT NULL,
  `host` varchar(128) DEFAULT NULL,
  `private_key` text DEFAULT NULL,
  `passphrase` varchar(128) DEFAULT NULL,
  `password` varchar(128) DEFAULT NULL,
  `node` varchar(128),
  `port` int(5) DEFAULT NULL,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  UNIQUE(`proxy_id`)
);

INSERT OR IGNORE INTO "t_bat_proxy" ("id", "proxy_id", "user", "host", "private_key", "passphrase", "password", "node", "port", "create_time", "update_time") VALUES
('1', 'localhost', 'user1', 'host1:22', NULL, NULL, '123456', 'j2001', '18880', '2023-04-12 09:35:39+00:00', '2023-04-12 09:35:39+00:00'),
('2', '127.0.0.1', 'user2', 'host2:22', NULL, NULL, '123456', 'g0156', '8888', '2023-04-12 09:35:39+00:00', '2023-04-12 09:35:39+00:00');
