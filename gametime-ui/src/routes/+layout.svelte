<script>
  import '../app.css';
  import { BottomNav, BottomNavItem, DarkMode, ImagePlaceholder, Skeleton } from 'flowbite-svelte';
  import {
    HomeSolid,
    WalletSolid,
    AdjustmentsVerticalOutline,
    UserCircleSolid,
  } from 'flowbite-svelte-icons';
  import { page } from '$app/stores';
  $: activeUrl = $page.url.pathname;
</script>

<div class="dark h-full w-full">
  <div>
    <slot></slot>
  </div>

  <div >
    asdf
    <DarkMode />
    asdf
    <BottomNav color="primary" {activeUrl} position="absolute" class="grid-cols-4">
      <BottomNavItem btnName="Home">
        <HomeSolid
          class="mb-1 h-6 w-6 text-gray-500 group-hover:text-primary-600 dark:text-gray-400 dark:group-hover:text-primary-500"
        />
      </BottomNavItem>
      <BottomNavItem btnName="Wallet">
        <WalletSolid
          class="mb-1 h-6 w-6 text-gray-500 group-hover:text-primary-600 dark:text-gray-400 dark:group-hover:text-primary-500"
        />
      </BottomNavItem>
      <BottomNavItem btnName="Settings">
        <AdjustmentsVerticalOutline
          class="mb-1 h-6 w-6 text-gray-500 group-hover:text-primary-600 dark:text-gray-400 dark:group-hover:text-primary-500"
        />
      </BottomNavItem>
      <BottomNavItem btnName="Profile">
        <UserCircleSolid
          class="mb-1 h-6 w-6 text-gray-500 group-hover:text-primary-600 dark:text-gray-400 dark:group-hover:text-primary-500"
        />
      </BottomNavItem>
    </BottomNav>
  </div>
</div>
