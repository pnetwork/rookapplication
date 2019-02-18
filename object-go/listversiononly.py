import boto3
import boto.s3.connection

access_key = 'D5ZZ0BY9SERTGEU4BBJG'
secret_key = 'rLMJHlwOnpXOvqz93V0IyOfKTnnnU4UrHMc5lmUY'
#conn = boto3.connect_s3(
s3client = boto3.client(service_name='s3',
        aws_access_key_id = access_key,
        aws_secret_access_key = secret_key,
        #host = '172.18.136.213/bucketvv', port = 80,
        endpoint_url = "http://localhost",
        )

#print 'into data1'
# list all bucket
response = s3client.list_buckets()
for bucket in response['Buckets']:
    print "Listing owned buckets returns => {0} was created on {1}\n".format(bucket['Name'], bucket['CreationDate'])

# create a new bucket
#bucket_name="newbucket"
bucket_name="newbucketgo2"
#response = s3client.create_bucket(Bucket = bucket_name)
#print "Creating bucket {0} returns => {1}\n".format(bucket_name, response)

# list all versioning by a bucket or by a bucket with filename
#resp=s3client.list_object_versions(Bucket = bucket_name)
resp=s3client.list_object_versions(Bucket = bucket_name, KeyMarker="go-mynewdata.log")
print "show versioning %s"%resp
