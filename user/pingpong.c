//
// Created by 柯淇文 on 2022/7/11.
//

#include "kernel/types.h"
#include "kernel/stat.h"
#include "user/user.h"


#define BUF_SIZE 16
int main(int argc, char *argv[]) {

	int fd[2];
	char buf[BUF_SIZE];

	pipe(fd);

	int pid = fork();
	if(pid > 0){
		int curPid = getpid();
		write(fd[1],"ping",12);
		wait((int*)0);
		read(fd[0],buf,BUF_SIZE);
		printf("%d : received %s\n",curPid,buf);
	} else{
		int curPid = getpid();
		read(fd[0],buf,BUF_SIZE);
		printf("%d : received %s\n",curPid,buf);
		write(fd[1],"pong",12);
	}

	exit(0);
}
