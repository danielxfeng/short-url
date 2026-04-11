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
import { useState } from 'react';

interface LinkRowCompProps {
  link: LinkRes;
  removeLink: (code: string) => Promise<void>;
  restoreDeleted: (code: string) => Promise<void>;
  isPending: boolean;
}

export const LinkRowComp = ({ link, removeLink, restoreDeleted, isPending }: LinkRowCompProps) => {
  const [showDeleteBtn, setShowDeleteBtn] = useState(false);
  const operationHandler = link.is_deleted ? restoreDeleted : removeLink;

  return (
    <>
      <TableRow onClick={() => setShowDeleteBtn((prev) => !prev)} className='cursor-pointer'>
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
        <TableCell className={link.is_deleted ? 'line-through text-muted-foreground' : undefined}>
          {link.original_url}
        </TableCell>
        <TableCell
          className={cn('text-center', link.is_deleted && 'line-through text-muted-foreground')}
        >
          {link.clicks}
        </TableCell>
      </TableRow>
      {showDeleteBtn && (
        <TableRow>
          <TableCell colSpan={3} className='text-center'>
            <Button
              onClick={async () => {
                await operationHandler(link.code);
                setShowDeleteBtn(false);
              }}
              disabled={isPending}
              variant={link.is_deleted ? 'outline' : 'destructive'}
              size='xs'
              className='w-1/2'
            >
              {link.is_deleted ? 'Restore' : 'Delete'}
            </Button>
          </TableCell>
        </TableRow>
      )}
    </>
  );
};

interface LinkTableCompProps {
  data: LinkRes[] | undefined;
  hasNext: boolean;
  fetchNext: () => void;
  isFetching: boolean;
  isFetchingNext: boolean;
  removeLink: (code: string) => Promise<void>;
  restoreDeleted: (code: string) => Promise<void>;
  isPending: boolean;
}

export const LinkTableComp = ({
  data,
  hasNext,
  fetchNext,
  isFetching,
  isFetchingNext,
  removeLink,
  restoreDeleted,
  isPending,
}: LinkTableCompProps) => (
  <section>
    <h2 className='font-medium text-base mb-3'>Links</h2>
    <div className='border rounded-2xl p-2'>
      <Table className='table-fixed'>
        <TableCaption className='sr-only'>The list of links.</TableCaption>
        <TableHeader>
          <TableRow>
            <TableHead className='w-20 text-center'>Code</TableHead>
            <TableHead className='text-center'>Url</TableHead>
            <TableHead className='w-20 text-center'>Visited</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data?.map((link) => (
            <LinkRowComp
              key={link.id}
              link={link}
              removeLink={removeLink}
              restoreDeleted={restoreDeleted}
              isPending={isPending}
            />
          ))}
        </TableBody>
      </Table>
      {isFetching && !data?.length && (
        <div className='flex justify-center py-4'>
          <Spinner />
        </div>
      )}
    </div>
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
  const { removeLink, restoreDeleted, isPending } = useMutateLink();

  return (
    <LinkTableComp
      data={data}
      hasNext={hasNext}
      fetchNext={fetchNext}
      isFetching={isFetching}
      isFetchingNext={isFetchingNext}
      removeLink={removeLink}
      restoreDeleted={restoreDeleted}
      isPending={isPending}
    />
  );
};

export default LinksTable;
