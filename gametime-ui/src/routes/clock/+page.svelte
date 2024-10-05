<script>
    import ClockButton from "$lib/ClockButton.svelte";

    export let clockState;
    export let clockConfig;
</script>


<div class="w-screen container h-screen flex flex-col justify-between mxx-auto">
<div class="mx-auto mt-24">
    <ClockButton label="Opponent" seconds={33}/>
    </div>
    <div class="mb-24 mx-4">
    <ClockButton seconds={33}/>
    </div>
</div>
