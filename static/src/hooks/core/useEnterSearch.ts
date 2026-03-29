export function useEnterSearch() {
  const createEnterSearchHandler = (search: () => void | Promise<void>) => {
    return (event: KeyboardEvent) => {
      if (event.key !== 'Enter' || event.isComposing) return
      void search()
    }
  }

  return {
    createEnterSearchHandler
  }
}
