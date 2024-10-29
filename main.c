#include <inttypes.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "types.h"

#define xstr(x) #x
#define str(x) xstr(x)

#define PROC_SWAP_FILE "/proc/swaps"
#define PROC_MEMINFO_FILE "/proc/meminfo"
#define PROC_STAT_FILE "/proc/stat"
#define PROC_DISKSTATS_FILE "/proc/diskstats"
#define PROC_UPTIME_FILE "/proc/uptime"

#define STREAM_BUFFER_INITIAL_SIZE 1024
#define BUFFER_GROWTH_FACTOR 2
#define MAX_NAME_SIZE 255

static void proc_swaps(swap_info_struct_t *const info) {
  uint32_t current_stream_size = STREAM_BUFFER_INITIAL_SIZE;
  char *stream = malloc(current_stream_size * sizeof(char));

  char name[MAX_NAME_SIZE + 1];
  char type[MAX_NAME_SIZE + 1];
  uint32_t size;
  uint32_t used;
  int32_t priority;

  FILE *swaps = fopen(PROC_SWAP_FILE, "r");

  bool isPrint = false;
  int count = 0;

  for (;;) {
    int c = fgetc(swaps);
    if (c == EOF) {
      break;
    }
    if (c == '\n') {
      isPrint = true;
    }
    if (isPrint) {
      stream[count] = c;
      count++;
      if (count >= current_stream_size - 1) {
        char *tmp = realloc(stream, current_stream_size * BUFFER_GROWTH_FACTOR *
                                        sizeof(char));
        if (tmp == NULL) {
          fprintf(stderr, "Error: failed to realloc memory\n");
          exit(1);
        }
        stream = tmp;
        current_stream_size = current_stream_size * BUFFER_GROWTH_FACTOR;
      }
    }
  }
  stream[count] = '\0';

  int ret = sscanf(stream,
                   "%" str(MAX_NAME_SIZE) "s %" str(MAX_NAME_SIZE) "s %u %u %d",
                   name, type, &size, &used, &priority);

  if (ret != 5) {
    fprintf(stderr, "Error: failed to parse /proc/swaps\n");
    fclose(swaps);
    exit(1);
  }

  info->name = malloc(strlen(name) * sizeof(char));
  memcpy(info->name, name, strlen(name) + 1);
  info->type = malloc(strlen(type) * sizeof(char));
  memcpy(info->type, type, strlen(type) + 1);
  info->size = size;
  info->used = used;
  info->priority = priority;

  fclose(swaps);
  free(stream);
  return;
}

static char *read_generic_proc_file(char const *const file_path) {
  uint32_t current_stream_size = STREAM_BUFFER_INITIAL_SIZE;
  char *stream = malloc(current_stream_size * sizeof(char));

  FILE *mems = fopen(file_path, "r");

  int count = 0;

  for (;;) {
    int c = fgetc(mems);

    if (c == EOF) {
      break;
    }

    stream[count] = c;
    count++;
    if (count >= current_stream_size - 1) {
      char *tmp = realloc(stream, current_stream_size * BUFFER_GROWTH_FACTOR *
                                      sizeof(char));
      if (tmp == NULL) {
        fprintf(stderr, "Error: failed to realloc memory\n");
        fclose(mems);
        exit(1);
      }
      stream = tmp;
      current_stream_size =
          current_stream_size * BUFFER_GROWTH_FACTOR * sizeof(char);
    }
  }
  stream[count] = '\0';
  fclose(mems);

  return stream;
}

#pragma weak main
int main(int argc, char **argv) {
  swap_info_struct_t swap_info;
  proc_swaps(&swap_info);
  printf("name: %s\n", swap_info.name);
  printf("type: %s\n", swap_info.type);
  printf("size: %u\n", swap_info.size);
  printf("used: %u\n", swap_info.used);
  printf("priority: %d\n", swap_info.priority);
  free(swap_info.name);
  free(swap_info.type);

  char *meminfo_stream = read_generic_proc_file(PROC_MEMINFO_FILE);
  printf("%s\n", meminfo_stream);
  free(meminfo_stream);

  char *statinfo_stream = read_generic_proc_file(PROC_STAT_FILE);
  printf("%s\n", statinfo_stream);
  free(statinfo_stream);

  char *diskstat_stream = read_generic_proc_file(PROC_DISKSTATS_FILE);
  printf("%s\n", diskstat_stream);
  free(diskstat_stream);

  char *uptime_stream = read_generic_proc_file(PROC_UPTIME_FILE);
  printf("%s\n", uptime_stream);
  free(uptime_stream);

  return 0;
}
