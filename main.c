#include <inttypes.h>
#include <stdbool.h>
#include <stdio.h>

static void proc_swaps() {
  FILE *swaps = fopen("/proc/swaps", "r");
  char stream[1000];
  char name[256];
  char type[256];
  uint32_t size;
  uint32_t used;
  int32_t priority;

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
    }
  }
  stream[count] = '\0';

  int ret =
      sscanf(stream, "%s %s %u %u %d", name, type, &size, &used, &priority);

  printf("name: %s\n", name);
  printf("type: %s\n", type);
  printf("size: %u\n", size);
  printf("used: %u\n", used);
}

int main(int argc, char **argv) { proc_swaps(); }
