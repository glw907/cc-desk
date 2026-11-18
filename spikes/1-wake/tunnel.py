"""CONNECT tunnel on 127.0.0.1 for proof 1.

It logs only the host of each tunnel and never reads or keeps the traffic.
While the file named by --block exists, it closes every open tunnel and
refuses new ones, which cuts the connection the way a dropped link would.

Usage: python3 -I tunnel.py --port 18907 --log tunnel.log --block block
"""

import argparse
import asyncio
import os
import time

open_tunnels = set()


def stamp():
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


async def pipe(reader, writer):
    try:
        while data := await reader.read(65536):
            writer.write(data)
            await writer.drain()
    except (ConnectionError, asyncio.CancelledError):
        pass
    finally:
        writer.close()


async def handle(reader, writer, args):
    head = await reader.readline()
    parts = head.decode("latin-1").split()
    while (await reader.readline()) not in (b"\r\n", b"\n", b""):
        pass
    if len(parts) < 2 or parts[0] != "CONNECT":
        writer.close()
        return
    host, _, port = parts[1].rpartition(":")
    if os.path.exists(args.block):
        with open(args.log, "a") as log:
            log.write(f"{stamp()} refused {host}\n")
        writer.write(b"HTTP/1.1 503 Blocked\r\n\r\n")
        writer.close()
        return
    try:
        up_r, up_w = await asyncio.open_connection(host, int(port))
    except OSError:
        writer.write(b"HTTP/1.1 502 Bad Gateway\r\n\r\n")
        writer.close()
        return
    with open(args.log, "a") as log:
        log.write(f"{stamp()} tunnel {host}\n")
    writer.write(b"HTTP/1.1 200 Connection Established\r\n\r\n")
    await writer.drain()
    tasks = [asyncio.ensure_future(pipe(reader, up_w)),
             asyncio.ensure_future(pipe(up_r, writer))]
    entry = (tasks, up_w, writer)
    open_tunnels.add(entry)
    await asyncio.wait(tasks)
    open_tunnels.discard(entry)


async def watch_block(args):
    while True:
        await asyncio.sleep(1)
        if os.path.exists(args.block):
            for tasks, up_w, writer in list(open_tunnels):
                for t in tasks:
                    t.cancel()
                up_w.close()
                writer.close()
                open_tunnels.discard((tasks, up_w, writer))


async def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--port", type=int, required=True)
    ap.add_argument("--log", required=True)
    ap.add_argument("--block", required=True)
    args = ap.parse_args()

    async def cb(r, w):
        await handle(r, w, args)

    server = await asyncio.start_server(cb, "127.0.0.1", args.port)
    asyncio.ensure_future(watch_block(args))
    async with server:
        await server.serve_forever()


asyncio.run(main())
