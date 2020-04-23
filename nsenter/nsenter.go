package nsenter

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>


// this function will auto execute when this package be imported
// will be executed when this process start
__attribute__((constructor)) void enter_namespace(void) {
	// get mydocker pid
	char* mydocker_pid = getenv("mydocker_pid");
	if (mydocker_pid) {
		fprintf(stdout, "got mydocker_pid=%s\n", mydocker_pid);
	} else {
		fprintf(stdout, "missing mydocker_pid env skip nsenter\n");
		return;
	}

	// get mydocker cmd
	char* mydocker_cmd = getenv("mydocker_cmd");
	if (mydocker_cmd) {
		fprintf(stdout, "got mydocker_cmd=%s\n", mydocker_cmd);
	} else {
		fprintf(stdout, "missing mydocker_cmd env skip nsenter\n");
		return;
	}

	int i;
	char nspath[1024];
	char* namespaces[] = {"ipc", "uts", "net", "pid", "mnt"};

	// enter namespace
	for (i = 0; i < 5; i++) {
		// path example: /proc/pid/ns/ipc
		sprintf(nspath, "/proc/%s/ns/%s", mydocker_pid, namespaces[i]);

		int fd = open(nspath, O_RDONLY);

		if (-1 == setns(fd, 0)) {
			fprintf(stderr, "setns on %s namespace faileds: %s\n", namespaces[i], strerror(errno));
		} else {
			fprintf(stdout, "setns on %s namespace successed\n", namespaces[i]);
		}
		close(fd);
	}

	// exec cmd
	int res = system(mydocker_cmd);
	//exit
	exit(0);
	return;
}
*/
import "C"
