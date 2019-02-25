# y: bytes of object, t: thread of concurrency, l: loop of time(measure time)
threads=1
loops=1
bucket="s3-benchmark"
access_key="UEY22BDBH9I7AMOY11VH"
secret_key="fU6W9v88QRWo4FxpSmmB7zygaMN5axsfSloXmt4I"
target="127.0.0.1:80"
objkey="objkey"
filesize=500000000
loopsleeptime=0
#while :
#do 
 ./s3-bench -ls $loopsleeptime -y $filesize -k $objkey -t $threads -l $loops  -a $access_key -b $bucket -s $secret_key -u $target 
 sleep 3
# ./s3-bench -ls $loopsleeptime -y $filesize -k $objkey -t $threads -l $loops  -a $access_key -b $bucket -s $secret_key -u $target -c true 
#done
# delete immediately without lazy delete, do not use delete object api if you want to delete soom
#radosgw-admin bucket rm --bucket=s3-benchmark --purge-objects --bypass-gc

