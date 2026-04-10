import useLinks from '@/hooks/useLinks';
import useMutateLink from '@/hooks/useMutateLink';

import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Spinner } from '../ui/spinner';
import type { LinkRes } from '@/schemas/schemas';
import { cn } from '@/lib/utils';

interface LinkTableCompProps {
  data: LinkRes[] | undefined;
  hasNext: boolean;
  fetchNext: () => void;
  isFetching: boolean;
  isFetchingNext: boolean;
  removeLink: (code: string) => Promise<void>;
  isPending: boolean;
}

export const LinkTableComp = ({
  data,
  hasNext,
  fetchNext,
  isFetching,
  isFetchingNext,
  removeLink,
  isPending,
}: LinkTableCompProps) => (
  <section>
    <h2>Links</h2>
    <Table className='table-fixed'>
      <TableCaption className='sr-only'>The list of links.</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead className='w-20 text-center'>Code</TableHead>
          <TableHead className='text-center'>Url</TableHead>
          <TableHead className='w-20 text-center'>Visited</TableHead>
          <TableHead className='w-20 text-center'>Operation</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data?.map((link) => (
          <TableRow key={link.id}>
            <TableCell
              className={cn('text-center', link.is_deleted && 'line-through text-muted-foreground')}
            >
              <a
                href={`/${link.code}`}
                target='_blank'
                rel='noopener noreferrer'
                className='text-primary hover:underline'
              >
                {link.code}
              </a>
            </TableCell>
            <TableCell
              className={link.is_deleted ? 'line-through text-muted-foreground' : undefined}
            >
              <a
                href={link.original_url}
                target='_blank'
                rel='noopener noreferrer'
                className='block max-w-full truncate text-primary hover:underline'
              >
                {link.original_url}
              </a>
            </TableCell>
            <TableCell
              className={cn('text-center', link.is_deleted && 'line-through text-muted-foreground')}
            >
              {link.clicks}
            </TableCell>
            <TableCell className='text-center'>
              <Button
                onClick={() => void removeLink(link.code)}
                disabled={isPending || link.is_deleted}
                variant='destructive'
                size='xs'
              >
                Delete
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
    {isFetching && !data?.length && (
      <div className='flex justify-center py-4'>
        <Spinner />
      </div>
    )}
    {hasNext && (
      <Button
        onClick={() => void fetchNext()}
        disabled={isFetchingNext}
        variant='outline'
        className='w-full'
      >
        {isFetchingNext ? <Spinner /> : 'Load more'}
      </Button>
    )}
  </section>
);

const LinksTable = () => {
  const { data, hasNext, fetchNext, isFetching, isFetchingNext } = useLinks();
  const { removeLink, isPending } = useMutateLink();

  return (
    <LinkTableComp
      data={data}
      hasNext={hasNext}
      fetchNext={fetchNext}
      isFetching={isFetching}
      isFetchingNext={isFetchingNext}
      removeLink={removeLink}
      isPending={isPending}
    />
  );
};

export default LinksTable;
