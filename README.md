cmattoon/dockerenv
==================

Retrieves values from Docker environment variables

The primary use case for this was inspecting TLS certificates in Docker environment variables, which looks something like this:

Verify a TLS cert in a container `abc123` with PEM data in `MYAPP_TLS_CRT` and `MYAPP_TLS_KEY`:

```
$ ./dockerenv --container-id abc123 tls verify --cert MYAPP_TLS_CRT --key MYAPP_TLS_KEY
YYYY/MM/DD HH:MM:SS Loaded X509KeyPair with 1 certs
YYYY/MM/DD HH:MM:SS Certificate 0 (CA: false)
YYYY/MM/DD HH:MM:SS =========================
YYYY/MM/DD HH:MM:SS 	Subject          : CN=myapp-tls-certificate-subject
YYYY/MM/DD HH:MM:SS 	Subject Key Id   :
YYYY/MM/DD HH:MM:SS
YYYY/MM/DD HH:MM:SS 	Issuer           : CN=myapp-selfsigned-ca
YYYY/MM/DD HH:MM:SS 	Authority Key Id :
YYYY/MM/DD HH:MM:SS
YYYY/MM/DD HH:MM:SS 	Not Before : YYYY-MM-DD HH:MM:SS +0000 UTC   (1 month ago)
YYYY/MM/DD HH:MM:SS 	Not After  : YYYY-MM-DD HH:MM:SS +0000 UTC   (5 months from now)
YYYY/MM/DD HH:MM:SS
...
```

To save the value of `MYAPP_DATABASE_HOST` in container `abc123` to the local var `DB_PASSWORD`:

    $ DB_PASSWORD=$(dockerenv -c abc123 -v MYAPP_DATABASE_PASS)

