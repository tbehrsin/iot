# iot

## Preparing a release

```
ENV=development make debian
ls dist/*.img
cd debian/
./run-diff.sh debian-previous-release.img debian-current-release.img
```

`debian-current-release.tar.gz` file will be created inside `dist/` folder.
