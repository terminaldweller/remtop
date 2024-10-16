#ifndef __types_h__
#define __types_h__

#include <inttypes.h>

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

typedef struct stat_info_struct_t {
  uint32_t cpu_user;
  uint32_t cpu_nice;
  uint32_t cpu_system;
  uint32_t cpu_idle;
  uint32_t cpu_iowait;
  uint32_t cpu_irq;
  uint32_t cpu_softirq;
  uint32_t cpu_steal;
  uint32_t cpu_guest;
  uint32_t cpu_guest_nice;
} stat_info_struct_t;

typedef struct diskstat_struct_t {
  uint32_t major;
  uint32_t minor;
  char *name;
  uint32_t reads_completed;
  uint32_t reads_merged;
  uint32_t sectors_read;
  uint32_t time_spent_reading;
  uint32_t writes_completed;
  uint32_t writes_merged;
  uint32_t sectors_written;
  uint32_t time_spent_writing;
  uint32_t io_in_progress;
  uint32_t time_spent_doing_io;
  uint32_t weighted_time_spent_doing_io;
} diskstat_struct_t;

typedef struct uptime_struct_t {
  uint32_t uptime;
  uint32_t idle_time;
} uptime_struct_t;

#endif // __types_h__
