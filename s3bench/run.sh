# y: bytes of object, t: thread of concurrency, l: loop of time(measure time)
threads=3
loops=1
bucket="s3-benchlargefile"
access_key="N3UDX9XCXPCZ1WN55E7S"
secret_key="xbm8X8oedVGSvzlthEbgPSOuSjtPRjaG8eA62B0K"
target="rook-ceph-rgw-my-store.rook-ceph"
objkey="objkey"
filesize=500000000
loopsleeptime=1
while :
do
 ./s3-bench -ls $loopsleeptime -y $filesize -k $objkey -t $threads -l $loops  -a $access_key -b $bucket -s $secret_key -u $target
 sleep 3
# ./s3-bench -ls $loopsleeptime -y $filesize -k $objkey -t $threads -l $loops  -a $access_key -b $bucket -s $secret_key -u $target -c true
radosgw-admin bucket rm --bucket=$bucket --purge-objects --bypass-gc
done
# delete immediately without lazy delete, do not use delete object api if you want to delete soom
