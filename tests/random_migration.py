
import time
import random
from redistrib import command
while True:
    slot = random.randint(0, 16383)
    src = random.randint(7001, 7003)
    dest = random.randint(7001, 7003)
    if src == dest:
        continue
    s = time.time()
    try:
        command.migrate_slots('127.0.0.1', src, '127.0.0.1', dest, [slot,])
    except Exception as e:
        print e
        pass
    else:
        e = time.time()
        print("migrate slot %d 127.0.0.1:%d -> 127.0.0.1:%d time: %f" % (slot, src, dest, e-s))

    time.sleep(1)

