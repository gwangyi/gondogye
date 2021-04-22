#ifndef __DHT_H__
#define __DHT_H__

#include <stdint.h>

struct dht_sensor {
  int pin;
  int max_counter;
  uint64_t last_timestamp;
  int last_level;
  int loglevel;
};

void dht_init(struct dht_sensor * dht, int pin);
void dht_set_loglevel(struct dht_sensor * dht, int loglevel);
int dht_read(struct dht_sensor * dht, uint32_t * output);

#endif

// vim: ft=c
