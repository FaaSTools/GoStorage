# GoStorage

## Info

This project serves as a unified abstraction layer for cloud storage services and can be used within any **Go** programs to access the storage services from the following providers:

* Google Cloud Storage

* Amazon S3

## Requirements

_aws-credentials.yaml:_

````yaml
labRole: "<FUNCTION_USER_GROUP>"
aws_access_key_id: "<ACCESS_KEY_ID>"
aws_secret_access_key: "<SECRET_ACCESS_KEY>"
aws_session_token: "<SESSION_TOKEN>"
````

Info: When using this library in combination with the _AWSAcademy_ course **labRole** will most likely be _RoleUser_.

_gcp-credentials.yaml:_

````yaml
{
  "type": "service_account",
  "project_id": "<PROJECT_ID>",
  "private_key_id": "<PRIVATE_KEY_ID>",
  "private_key": "-----BEGIN PRIVATE KEY-----<PRIVATE_KEY>-----END PRIVATE KEY-----\n"
}
````

For more information how to retrieve the information needed for this file, see: [Google Cloud](https://cloud.google.com/iam/docs/creating-managing-service-accounts)


## Project Structure

The structure for a project using *GoStorage* should look something like this.

``` shell
.
├── aws-credentials.yaml
├── gcp-credentials.yaml
├── code
│   ├── ...
```