//
// Created by 柯淇文 on 2022/7/11.
//

#include "kernel/types.h"
#include "kernel/stat.h"
#include "user/user.h"
#include "kernel/fs.h"
#include "kernel/param.h"

#define BUF_SIZE 16
int main(int argc, char *argv[]) {

	sleep(20);
	//echo hello | xargs echo by
	// by hello

	char buf[BUF_SIZE];

	//怎么获取前一个的输入
	read(0,buf,BUF_SIZE);

	//获取自己的命令行参数

	char *xargv[MAXARG];
	int xargc = 0;
	for (int i = 1; i < argc ; ++i) {
		xargv[xargc++] = argv[i];
	}

	char *p = buf;//需要这个p指针将前一个的输入补充到 xagrv里面
	for (int i = 0; i < BUF_SIZE; ++i) {
		if(buf[i] == '\n'){
			int pid = fork();
			if(pid > 0){
				p = &buf[i+1];
				wait(0);
			} else{
				buf[i] = 0;
				xargv[xargc++] = p;
				xargv[xargc++] = 0;

				exec(xargv[0],xargv);
				exit(0);
			}
		}
	}
	wait(0);
	exit(0);
}
