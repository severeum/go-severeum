/*
  This file is part of sevash.

  sevash is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  sevash is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with sevash.  If not, see <http://www.gnu.org/licenses/>.
*/

/** @file sevash.h
* @date 2015
*/
#pragma once

#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include <stddef.h>
#include "compiler.h"

#define SEVASH_REVISION 23
#define SEVASH_DATASET_BYTES_INIT 1073741824U // 2**30
#define SEVASH_DATASET_BYTES_GROWTH 8388608U  // 2**23
#define SEVASH_CACHE_BYTES_INIT 1073741824U // 2**24
#define SEVASH_CACHE_BYTES_GROWTH 131072U  // 2**17
#define SEVASH_EPOCH_LENGTH 30000U
#define SEVASH_MIX_BYTES 128
#define SEVASH_HASH_BYTES 64
#define SEVASH_DATASET_PARENTS 256
#define SEVASH_CACHE_ROUNDS 3
#define SEVASH_ACCESSES 64
#define SEVASH_DAG_MAGIC_NUM_SIZE 8
#define SEVASH_DAG_MAGIC_NUM 0xFEE1DEADBADDCAFE

#ifdef __cplusplus
extern "C" {
#endif

/// Type of a seedhash/blockhash e.t.c.
typedef struct sevash_h256 { uint8_t b[32]; } sevash_h256_t;

// convenience macro to statically initialize an h256_t
// usage:
// sevash_h256_t a = sevash_h256_static_init(1, 2, 3, ... )
// have to provide all 32 values. If you don't provide all the rest
// will simply be unitialized (not guranteed to be 0)
#define sevash_h256_static_init(...)			\
	{ {__VA_ARGS__} }

struct sevash_light;
typedef struct sevash_light* sevash_light_t;
struct sevash_full;
typedef struct sevash_full* sevash_full_t;
typedef int(*sevash_callback_t)(unsigned);

typedef struct sevash_return_value {
	sevash_h256_t result;
	sevash_h256_t mix_hash;
	bool success;
} sevash_return_value_t;

/**
 * Allocate and initialize a new sevash_light handler
 *
 * @param block_number   The block number for which to create the handler
 * @return               Newly allocated sevash_light handler or NULL in case of
 *                       ERRNOMEM or invalid parameters used for @ref sevash_compute_cache_nodes()
 */
sevash_light_t sevash_light_new(uint64_t block_number);
/**
 * Frees a previously allocated sevash_light handler
 * @param light        The light handler to free
 */
void sevash_light_delete(sevash_light_t light);
/**
 * Calculate the light client data
 *
 * @param light          The light client handler
 * @param header_hash    The header hash to pack into the mix
 * @param nonce          The nonce to pack into the mix
 * @return               an object of sevash_return_value_t holding the return values
 */
sevash_return_value_t sevash_light_compute(
	sevash_light_t light,
	sevash_h256_t const header_hash,
	uint64_t nonce
);

/**
 * Allocate and initialize a new sevash_full handler
 *
 * @param light         The light handler containing the cache.
 * @param callback      A callback function with signature of @ref sevash_callback_t
 *                      It accepts an unsigned with which a progress of DAG calculation
 *                      can be displayed. If all goes well the callback should return 0.
 *                      If a non-zero value is returned then DAG generation will stop.
 *                      Be advised. A progress value of 100 means that DAG creation is
 *                      almost complete and that this function will soon return succesfully.
 *                      It does not mean that the function has already had a succesfull return.
 * @return              Newly allocated sevash_full handler or NULL in case of
 *                      ERRNOMEM or invalid parameters used for @ref sevash_compute_full_data()
 */
sevash_full_t sevash_full_new(sevash_light_t light, sevash_callback_t callback);

/**
 * Frees a previously allocated sevash_full handler
 * @param full    The light handler to free
 */
void sevash_full_delete(sevash_full_t full);
/**
 * Calculate the full client data
 *
 * @param full           The full client handler
 * @param header_hash    The header hash to pack into the mix
 * @param nonce          The nonce to pack into the mix
 * @return               An object of sevash_return_value to hold the return value
 */
sevash_return_value_t sevash_full_compute(
	sevash_full_t full,
	sevash_h256_t const header_hash,
	uint64_t nonce
);
/**
 * Get a pointer to the full DAG data
 */
void const* sevash_full_dag(sevash_full_t full);
/**
 * Get the size of the DAG data
 */
uint64_t sevash_full_dag_size(sevash_full_t full);

/**
 * Calculate the seedhash for a given block number
 */
sevash_h256_t sevash_get_seedhash(uint64_t block_number);

#ifdef __cplusplus
}
#endif
