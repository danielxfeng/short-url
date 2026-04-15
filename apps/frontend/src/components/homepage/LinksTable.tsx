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
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import type { LinkRes } from '@/schemas/schemas';
import { cn } from '@/lib/utils';
import { useState, type Dispatch } from 'react';
import { toast } from 'sonner';
import useInfiniteScroll from '@/hooks/useInfiniteScroll';

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
        {/* Code */}
        <TableCell
          className={cn('text-center', link.isDeleted && 'line-through text-muted-foreground')}
        >
          <Tooltip>
            <TooltipTrigger asChild>
              <a
                href={`/${link.code}`}
                target='_blank'
                rel='noopener'
                className='hover:underline truncate block'
                onClick={(e) => e.stopPropagation()}
              >
                {link.code}
              </a>
            </TooltipTrigger>
            <TooltipContent>
              <p>{link.originalUrl}</p>
            </TooltipContent>
          </Tooltip>
        </TableCell>

        {/* Clicks */}
        <TableCell
          className={cn('text-center', link.isDeleted && 'line-through text-muted-foreground')}
        >
          {link.clicks}
        </TableCell>

        {/* Note */}
        <TableCell
          className={cn('text-center', link.isDeleted && 'line-through text-muted-foreground')}
        >
          <div className='truncate'>{link.note}</div>
        </TableCell>
      </TableRow>

      {/* Delete button row */}
      {showDeleteBtn && (
        <TableRow>
          <TableCell colSpan={3} className='text-center'>
            <div className='flex items-center justify-center gap-4'>
              {/* Delete/Restore button */}
              <Button
                className='px-4'
                onClick={(e) => {
                  e.stopPropagation();
                  void operationHandler(link.isDeleted ? 'restore' : 'remove');
                }}
                disabled={isPending}
                variant={link.isDeleted ? 'outline' : 'destructive'}
                size='xs'
              >
                {isPending ? <Spinner /> : link.isDeleted ? 'Restore' : 'Delete'}
              </Button>

              {/* Permanently delete button, only show when the link is already soft deleted */}
              {link.isDeleted && (
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant='destructive'
                      className='px-4'
                      size='xs'
                      onClick={(e) => void e.stopPropagation()}
                      disabled={isPending}
                    >
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
  isFetching: boolean;
  isFetchingNext: boolean;
  removeLink: (code: string) => Promise<void>;
  restoreDeleted: (code: string) => Promise<void>;
  permanentlyDelete: (code: string) => Promise<void>;
  isPending: boolean;
  loadMoreRef: Dispatch<React.SetStateAction<HTMLDivElement | null>>;
}

export const LinkTableComp = ({
  data,
  hasNext,
  isFetching,
  isFetchingNext,
  removeLink,
  restoreDeleted,
  permanentlyDelete,
  isPending,
  loadMoreRef,
}: LinkTableCompProps) => (
  <section className='flex flex-col gap-4'>
    <h2 className='text-base font-medium tracking-tight'>Links</h2>
    <div className='border rounded-2xl p-2 bg-background'>
      <Table className='table-fixed'>
        <TableCaption className='sr-only'>The list of links.</TableCaption>
        <TableHeader>
          <TableRow>
            <TableHead className='w-28 text-center'>Code</TableHead>
            <TableHead className='w-20 text-center'>Clicks</TableHead>
            <TableHead className='text-center'>Note</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {!isFetching && data?.length === 0 ? (
            <TableRow>
              <TableCell colSpan={3} className='text-center'>
                No links found
              </TableCell>
            </TableRow>
          ) : (
            data?.map((link) => (
              <LinkRowComp
                key={link.id}
                link={link}
                removeLink={removeLink}
                restoreDeleted={restoreDeleted}
                permanentlyDelete={permanentlyDelete}
                isPending={isPending}
              />
            ))
          )}
        </TableBody>
      </Table>

      {/* Spinner for initial fetch */}
      {isFetching && !data?.length && (
        <div className='flex justify-center py-4'>
          <Spinner />
        </div>
      )}
    </div>

    {/* Load more */}
    {hasNext && (
      <div ref={loadMoreRef} className='flex w-full min-h-10 items-center justify-center'>
        {isFetchingNext && <Spinner />}
      </div>
    )}
  </section>
);

const LinksTable = () => {
  const { data, hasNext, fetchNext, isFetching, isFetchingNext } = useLinks();
  const { removeLink, restoreDeleted, permanentlyDelete, isPending } = useMutateLink();
  const loadMoreRef = useInfiniteScroll(hasNext, isFetchingNext, fetchNext);

  return (
    <LinkTableComp
      data={data}
      hasNext={hasNext}
      isFetching={isFetching}
      isFetchingNext={isFetchingNext}
      removeLink={removeLink}
      restoreDeleted={restoreDeleted}
      permanentlyDelete={permanentlyDelete}
      isPending={isPending}
      loadMoreRef={loadMoreRef}
    />
  );
};

export default LinksTable;
