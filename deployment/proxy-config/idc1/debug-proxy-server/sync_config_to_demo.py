#!/usr/bin/env python
import re


def changeIP(line):
    """
        Replace the server ip to a fake server ip for demo
        e.g., server idc52 10.0.0.52:80
        ->    server idc52 172.17.0.1:10052
    """
    regex = r"server idc(\d+) ([\d|:|\.]+)"
    match = re.search(regex, line)
    if match:
        ip = match.group(1)
        old = match.group(2)
        new = "172.17.0.1:100{}".format(ip)
        return line.replace(old, new)
    else:
        return line


def genDemoConfig(infile, outfile):
    out = open(outfile, "w")
    with open(infile) as f:
        lines = f.readlines()
        for line in lines:
            out.write(changeIP(line))
    out.close()

if __name__ == '__main__':
    inFn = "../haproxy-idc1/haproxy.cfg"
    outFn = "haproxy.demo.cfg"
    print "# {0} -> {1}".format(inFn, outFn)
    genDemoConfig(inFn, outFn)
