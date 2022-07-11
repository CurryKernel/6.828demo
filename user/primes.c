//
// Created by 柯淇文 on 2022/7/11.
//
#include "kernel/types.h"
#include "kernel/stat.h"
#include "user/user.h"

#define BUF_SIZE 35


void prime(int read,int write){

	char buf[BUF_SIZE];
	int tmp = 0;
	read(read,buf,BUF_SIZE);
	for (int i = 0; i < BUF_SIZE; ++i) {
		if(buf[i] == 1){
			tmp = i;
			break;
		}
	}

	if(tmp == 0){
		exit(0);
	}

	printf("prime %d\n",tmp);

	for (int i = 0; i < BUF_SIZE; ++i) {
		if(i % tmp == 0){
			buf[i] = 0;
		}
	}

	int pid = fork();
	if(pid > 0){
		write(write,buf,BUF_SIZE);
	} else{
		prime(read,write);
	}
}
int main(int argc, char *argv[]) {

	int fd[2];
	pipe(fd);

	char buf[BUF_SIZE];

	for(int i = 2;i < BUF_SIZE;i++){
		buf[i] = 1;
	}//标记假设都是素数

	int pid = fork();
	if(pid > 0){
		buf[0] = 0;
		buf[1] = 1;
		write(fd[1],buf,BUF_SIZE);
		wait(0);
	} else{
		prime(fd[0],fd[1]);
		wait(0);
	}

	exit(0);
//	return 0;
}
