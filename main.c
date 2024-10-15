#include <inttypes.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct swap_info_struct_t {
  char *name;
  char *type;
  uint32_t size;
  uint32_t used;
  int32_t priority;
} swap_info_struct_t;

typedef struct mem_info_struct_t {
  uint32_t total;
  uint32_t free;
  uint32_t available;
  uint32_t buffers;
  uint32_t cached;
  uint32_t swap_total;
  uint32_t swap_free;
} mem_info_struct_t;

#define xstr(x) #x
#define str(x) xstr(x)

#define PROC_SWAP_FILE "/proc/swaps"
#define PROC_MEMINFO_FILE "/proc/meminfo"

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
        stream = realloc(stream, current_stream_size * BUFFER_GROWTH_FACTOR *
                                     sizeof(char));
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
    exit(1);
  }

  info->name = malloc(strlen(name) * sizeof(char));
  memcpy(info->name, name, strlen(name) + 1);
  info->type = malloc(strlen(type) * sizeof(char));
  memcpy(info->type, type, strlen(type) + 1);
  info->size = size;
  info->used = used;
  info->priority = priority;

  free(stream);
}

static void proc_meminfo(mem_info_struct_t *const info) {
  uint32_t current_stream_size = STREAM_BUFFER_INITIAL_SIZE;
  char *stream = malloc(current_stream_size * sizeof(char));

  FILE *mems = fopen(PROC_MEMINFO_FILE, "r");

  int count = 0;

  for (;;) {
    int c = fgetc(mems);

    if (c == EOF) {
      break;
    }

    stream[count] = c;
    count++;
    if (count >= current_stream_size - 1) {
      stream = realloc(stream, current_stream_size * BUFFER_GROWTH_FACTOR *
                                   sizeof(char));
      current_stream_size = current_stream_size * BUFFER_GROWTH_FACTOR;
    }
  }
  stream[count] = '\0';

  printf("%s\n", stream);

  free(stream);
}

int main(int argc, char **argv) {
  swap_info_struct_t swap_info;
  proc_swaps(&swap_info);

  printf("name: %s\n", swap_info.name);
  printf("type: %s\n", swap_info.type);
  printf("size: %u\n", swap_info.size);
  printf("used: %u\n", swap_info.used);
  printf("priority: %d\n", swap_info.priority);

  mem_info_struct_t mem_info;
  proc_meminfo(&mem_info);

  return 0;
}
