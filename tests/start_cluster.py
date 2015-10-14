import os
import sys
import logging
import time
from redistrib import command

PORTS = [
    7001,
    7002,
    7003,
    7004,
    7005,
    7006,
]
assert len(PORTS) % 2 == 0

cluster = "cluster"
tmpl_file = "redis.conf.tmpl"


def mkdir_p(dir):
    os.system("mkdir -p %s" % dir)


def start_servers():
    logging.info("start servers")
    tmpl = open(tmpl_file, 'r').read()
    for port in PORTS:
        dir = "%s/%s" % (cluster, port)
        mkdir_p(dir)
        with open(os.path.join(dir, "redis.conf"), 'w') as fp:
            fp.write(tmpl.format(PORT=port))

        cmd = "cd %s; redis-server redis.conf" % dir
        os.system(cmd)


def start_cluster():
    logging.info("start cluster")
    servers = [('127.0.0.1', port) for port in PORTS]
    half = len(servers) / 2
    command.start_cluster_on_multi(servers[0:half])
    time.sleep(5)
    for i in range(half):
        command.replicate("127.0.0.1", PORTS[i], "127.0.0.1", PORTS[i + half])


if os.path.exists(cluster):
    start_servers()
    logging.warn("cluster dir has already exists")
    sys.exit(0)
else:
    start_servers()
    start_cluster()
logging.info("done")









