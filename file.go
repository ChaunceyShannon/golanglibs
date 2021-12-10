package golanglibs

import (
	"io"
	"os"
)

type fileStruct struct {
	filePath string
}

func File(filePath string) *fileStruct {
	return &fileStruct{filePath: filePath}
}

type fileTimeStruct struct {
	ctime int64
	mtime int64
	atime int64
}

func (f *fileStruct) Time() *fileTimeStruct {
	fi, err := os.Stat(f.filePath)
	panicerr(err)
	mtime := fi.ModTime().Unix()
	// stat := fi.Sys().(*syscall.Stat_t)
	// ctime := time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec)).Unix()
	// atime := time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec)).Unix()

	return &fileTimeStruct{
		mtime: mtime,
		// ctime: ctime,
		// atime: atime,
	}
}

// Get file details
func (f *fileStruct) Stat() os.FileInfo {
	ff, err := os.Stat(f.filePath)
	panicerr(err)
	return ff
}

// Get file size
func (f *fileStruct) Size() int64 {
	info, err := os.Stat(f.filePath)
	panicerr(err)
	return info.Size()
}

// Touch a file like touch command
func (f *fileStruct) Touch() {
	fd, err := os.OpenFile(f.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	panicerr(err)
	fd.Close()
}

func (f *fileStruct) Chmod(mode os.FileMode) bool {
	return os.Chmod(f.filePath, mode) == nil
}

func (f *fileStruct) Chown(uid, gid int) bool {
	return os.Chown(f.filePath, uid, gid) == nil
}

func (f *fileStruct) Mtime() int64 {
	fd, err := os.Open(f.filePath)
	panicerr(err)
	defer fd.Close()
	fileinfo, err := fd.Stat()
	panicerr(err)
	return fileinfo.ModTime().Unix()
}

func (f *fileStruct) Unlink() {
	err := os.RemoveAll(f.filePath)
	panicerr(err)
}

func (f *fileStruct) Copy(dest string) {
	fd1, err := os.Open(f.filePath)
	panicerr(err)
	defer fd1.Close()
	fd2, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0644)
	panicerr(err)
	defer fd2.Close()
	_, err = io.Copy(fd2, fd1)
	panicerr(err)
}

func (f *fileStruct) Move(newPosition string) {
	err := os.Rename(f.filePath, newPosition)
	panicerr(err)
	f.filePath = newPosition
}
