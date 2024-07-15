# Battle of wits

A simple program that generates a video with your ip address and location and sends it.
This is not a "clone and run" program, you'll need to setup some things before running it.

## How to run

```bash
docker run -d \
  --restart always \
  -v $(pwd)/assets:/assets \
  -v volume:/tmp \
  -p 0.0.0.0:3000:3000 \
  -e IPINFO_TOKEN=${IPINFO_TOKEN} \
  --mount type=tmpfs,destination=/tmp,tmpfs-size=500m \
  --name website \
  ghcr.io/sipacid/wits:main
```
