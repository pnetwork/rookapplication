# y: bytes of object, t: thread of concurrency, l: loop of time(measure time)
threads=10
loops=10
bucket="s3-benchmark"
access_key="CXPS27F2AWRVHJGDVGXA"
secrets_key="FCBqA35AGMsG5bPWdD1mA6Nfl8yV82iSeA6K4Ca7"
target="127.0.0.1:80"
objkey="objkey"
filesize=20000000
./s3-bench -y $filesize -k $objkey -t $threads -l $loops  -a $access_key -b $bucket -s $secrets_key -u $target 

