# y: bytes of object, t: thread of concurrency, l: loop of time(measure time)
./s3-benchmark -y 200000 -t 10 -l 100 -a CXPS27F2AWRVHJGDVGXA -b s3-benchmark -s FCBqA35AGMsG5bPWdD1mA6Nfl8yV82iSeA6K4Ca7 -u 127.0.0.1:80 -z 1k -c true

