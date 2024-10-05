<script lang="ts">
  import { Card } from 'flowbite-svelte';

  export let seconds = 65;
  export let label: string | null = null;

  const formatter = Intl.NumberFormat("en-US", {
      minimumIntegerDigits: 2,
  })

  let displayMins = Math.floor(seconds/60);
  let displaySeconds = formatter.format(seconds%60);

  $: formattedTime = `${displayMins}:${displaySeconds}`
</script>


<Card class="px-20 py-10 text-center cursor-pointer bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 hover:dark:bg-gray-600">
  {#if label}
      <p class="mb-2">{label}</p>
  {/if}
  <h5 class="text-3xl text-gray-900 dark:text-white">{formattedTime}</h5>
</Card>


