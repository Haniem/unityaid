<template>
  <RouterView v-slot="{ Component, route }">
    <Transition name="page-slide" mode="out-in" appear>
      <component :is="Component" :key="route.fullPath" />
    </Transition>
  </RouterView>
</template>
