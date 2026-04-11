import { useEffect, useRef, useState } from 'react';

const useInfiniteScroll = (
  hasNext: boolean,
  isFetching: boolean,
  fetchNext: () => void | Promise<unknown>,
) => {
  const [node, setNode] = useState<HTMLDivElement | null>(null);
  const hasNextRef = useRef(hasNext);
  const fetchingRef = useRef(isFetching);
  const fetchNextRef = useRef(fetchNext);

  useEffect(() => {
    hasNextRef.current = hasNext;
    fetchingRef.current = isFetching;
    fetchNextRef.current = fetchNext;
  }, [hasNext, isFetching, fetchNext]);

  useEffect(() => {
    if (!node) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (!entry?.isIntersecting) return;
        if (!hasNextRef.current || fetchingRef.current) return;

        void fetchNextRef.current();
      },
      { root: null, rootMargin: '100px', threshold: 0 },
    );

    observer.observe(node);

    return () => observer.disconnect();
  }, [node]);

  return setNode;
};

export default useInfiniteScroll;
