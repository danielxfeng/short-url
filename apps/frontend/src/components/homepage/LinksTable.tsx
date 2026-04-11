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
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import type { LinkRes } from '@/schemas/schemas';
import { cn } from '@/lib/utils';
import { useState } from 'react';
import { toast } from 'sonner';

interface LinkRowCompProps {
  link: LinkRes;
  removeLink: (code: string) => Promise<void>;
  restoreDeleted: (code: string) => Promise<void>;
  permanentlyDelete: (code: string) => Promise<void>;
  isPending: boolean;
}

export const LinkRowComp = ({
  link,
  removeLink,
  restoreDeleted,
  permanentlyDelete,
  isPending,
}: LinkRowCompProps) => {
  const [showDeleteBtn, setShowDeleteBtn] = useState(false);
  const operationHandler = async (method: 'remove' | 'restore' | 'permanently delete') => {
    try {
      switch (method) {
        case 'remove':
          await removeLink(link.code);
          break;
        case 'restore':
          await restoreDeleted(link.code);
          break;
        case 'permanently delete':
          await permanentlyDelete(link.code);
          break;
      }
      setShowDeleteBtn(false);
      toast.success(`Link ${method}d successfully.`);
    } catch {
      toast.error(`Failed to ${method} the link. Please try again.`);
    }
  };

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
            className='hover:underline'
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
            <div className='flex items-center justify-center gap-4'>
              <Button
                onClick={() => void operationHandler(link.is_deleted ? 'restore' : 'remove')}
                disabled={isPending}
                variant={link.is_deleted ? 'outline' : 'destructive'}
                size='xs'
              >
                {isPending ? <Spinner /> : link.is_deleted ? 'Restore' : 'Delete'}
              </Button>

              {link.is_deleted && (
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button variant='destructive' size='xs'>
                      {isPending ? <Spinner /> : 'Permanently Delete'}
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                      <AlertDialogDescription>
                        This action cannot be undone. This will permanently delete the link with
                        code <strong>{link.code}</strong>.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel variant='outline'>Cancel</AlertDialogCancel>
                      <AlertDialogAction
                        variant='destructive'
                        disabled={isPending}
                        onClick={() => void operationHandler('permanently delete')}
                      >
                        {isPending ? <Spinner /> : 'Permanently Delete'}
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              )}
            </div>
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
  permanentlyDelete: (code: string) => Promise<void>;
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
  permanentlyDelete,
  isPending,
}: LinkTableCompProps) => (
  <section className='flex flex-col gap-4'>
    <h2 className='font-medium text-base'>Links</h2>
    <div className='border rounded-2xl p-2 bg-background'>
      <Table className='table-fixed'>
        <TableCaption className='sr-only'>The list of links.</TableCaption>
        <TableHeader>
          <TableRow>
            <TableHead className='w-20 text-center'>Code</TableHead>
            <TableHead className='text-center'>URL</TableHead>
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
              permanentlyDelete={permanentlyDelete}
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
  const { removeLink, restoreDeleted, permanentlyDelete, isPending } = useMutateLink();

  return (
    <LinkTableComp
      data={data}
      hasNext={hasNext}
      fetchNext={fetchNext}
      isFetching={isFetching}
      isFetchingNext={isFetchingNext}
      removeLink={removeLink}
      restoreDeleted={restoreDeleted}
      permanentlyDelete={permanentlyDelete}
      isPending={isPending}
    />
  );
};

export default LinksTable;
