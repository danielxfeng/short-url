import { infiniteQueryOptions, useInfiniteQuery, type InfiniteData } from '@tanstack/react-query';
import { getLinks } from '../services';
import type { LinksRes } from '@/schemas/schemas';

export const linksQueryOptions = () =>
  infiniteQueryOptions({
    queryKey: ['links'],
    queryFn: ({ pageParam }: { pageParam: number | undefined }) => getLinks(pageParam),
    initialPageParam: undefined,
    getNextPageParam: (lastPage: LinksRes) => (lastPage.has_more ? lastPage.cursor : undefined),
    select: (data: InfiniteData<LinksRes>) => data.pages.flatMap((page) => page.links),
  });

const useLinks = () => {
  const query = useInfiniteQuery(linksQueryOptions());

  return {
    data: query.data,
    hasNext: query.hasNextPage,
    fetchNext: query.fetchNextPage,
    isFetching: query.isFetching,
    isFetchingNext: query.isFetchingNextPage,
    isError: query.isError,
    error: query.error,
  };
};

export default useLinks;
