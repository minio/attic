## Update `simd` dependency

```sh
git remote add -f simd https://github.com/minio/simd.git
git fetch simd master
git subtree pull --prefix contrib/Simd simd master --squash
```
