import boto3
import boto.s3.connection

access_key = 'USB1A7W8W8RXZ7AWODB1'
secret_key = 'K8HpKFRy2KGLxr0ekH06EiWdnvzNBMqUHS3XHJv8'
#conn = boto3.connect_s3(
s3client = boto3.client(service_name='s3',
        aws_access_key_id = access_key,
        aws_secret_access_key = secret_key,
        #host = '172.18.136.213/bucketvv', port = 80,
        endpoint_url = "http://172.18.136.213",
        )

#print 'into data1'
# list all bucket
response = s3client.list_buckets()
for bucket in response['Buckets']:
    print "Listing owned buckets returns => {0} was created on {1}\n".format(bucket['Name'], bucket['CreationDate'])

# create a new bucket
bucket_name="newbucket"
response = s3client.create_bucket(Bucket = bucket_name)
print "Creating bucket {0} returns => {1}\n".format(bucket_name, response)

# list all versioning by a bucket or by a bucket with filename
#resp=s3client.list_object_versions(Bucket = bucket_name)
resp=s3client.list_object_versions(Bucket = bucket_name, KeyMarker="mynewdata.log")
print "show versioning %s"%resp
)
