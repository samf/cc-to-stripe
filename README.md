## Quick debug

`% m && ./cc-to-stripe -i -l foo.bar.com -p :8080`

## Run in production

`# docker run -d -p 443:443 -p 80:80 -v /autocert:/autocert samfalkner/cc-to-stripe [ -i (or any other CLI opts) ]`

## DigitalOcean Instructions

1. Choose an image that is docker friendly, preferably a 1-click-app for docker.

2. ssh root@ip-address

3. docker login

4. docker pull samfalkner/cc-to-stripe

5. mkdir /autocert

6. docker run ... (see above)

7. test with a non-vital domain

8. If all goes well, shut down docker, mount the volume, and restart docker.
