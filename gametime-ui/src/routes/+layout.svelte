<script>
  import '../app.css';
  import { BottomNav, BottomNavItem, DarkMode, ImagePlaceholder, Skeleton } from 'flowbite-svelte';
  import {
    HomeSolid,
    WalletSolid,
    AdjustmentsVerticalOutline,
    UserCircleSolid,
    ClockSolid,
  } from 'flowbite-svelte-icons';
  import { page } from '$app/stores';
  $: activeUrl = $page.url.pathname;
</script>

<DarkMode size="lg" class="absolute top-3 right-3"/>

<slot></slot>

<BottomNav {activeUrl} position="absolute" navType="card" classInner="grid-cols-3">
  <BottomNavItem btnName="Clock" href="/clock">
    <ClockSolid class="w-6 h-6 mb-1 text-gray-500 dark:text-gray-400 group-hover:text-primary-600 dark:group-hover:text-primary-500" />
  </BottomNavItem>
  <BottomNavItem btnName="Settings">
    <AdjustmentsVerticalOutline class="w-6 h-6 mb-1 text-gray-500 dark:text-gray-400 group-hover:text-primary-600 dark:group-hover:text-primary-500" />
  </BottomNavItem>
  <BottomNavItem btnName="Preferences">
    <UserCircleSolid class="w-6 h-6 mb-1 text-gray-500 dark:text-gray-400 group-hover:text-primary-600 dark:group-hover:text-primary-500" />
  </BottomNavItem>
</BottomNav>
