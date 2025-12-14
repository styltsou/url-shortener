import { useState, useRef, useEffect } from 'react'
import { Check, ChevronsUpDown, Plus, Tag as TagIcon } from 'lucide-react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Button } from '@/components/ui/button'
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList, CommandSeparator } from '@/components/ui/command'
import { Spinner } from '@/components/ui/spinner'
import { cn } from '@/lib/utils'
import type { Tag } from '@/types/url'

interface TagComboboxProps {
  availableTags: Tag[]
  selectedTags: Tag[]
  onTagSelect: (tag: Tag) => void
  onCreateTag: (tagName: string) => Promise<Tag>
  placeholder?: string
  className?: string
}

export function TagCombobox({
  availableTags,
  selectedTags,
  onTagSelect,
  onCreateTag,
  placeholder = 'Add a tag...',
  className,
}: TagComboboxProps) {
  const [open, setOpen] = useState(false)
  const [searchValue, setSearchValue] = useState('')
  const [isCreating, setIsCreating] = useState(false)

  // Filter available tags to exclude already selected tags
  // React Compiler automatically memoizes this computation
  const availableTagsForSelection = availableTags.filter(
    (tag) => !selectedTags.find((t) => t.id === tag.id)
  )

  // Check if we should show "create tag" option
  // React Compiler automatically memoizes this computation
  const searchLower = searchValue.toLowerCase().trim()
  const shouldShowCreateOption = searchLower
    ? !availableTags.some((tag) => tag.name.toLowerCase() === searchLower) &&
      !selectedTags.some((tag) => tag.name.toLowerCase() === searchLower)
    : false

  // Reset search when popover closes
  useEffect(() => {
    if (!open) {
      setSearchValue('')
    }
  }, [open])

  const handleCreateTag = async () => {
    const tagName = searchValue.trim()
    if (!tagName || isCreating) return

    setIsCreating(true)
    try {
      const newTag = await onCreateTag(tagName)
      onTagSelect(newTag)
      setSearchValue('')
      setOpen(false)
    } catch (error) {
      console.error('Failed to create tag:', error)
    } finally {
      setIsCreating(false)
    }
  }

  const handleTagSelect = (tag: Tag) => {
    onTagSelect(tag)
    setSearchValue('')
    setOpen(false)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && shouldShowCreateOption && searchValue.trim()) {
      e.preventDefault()
      handleCreateTag()
    } else if (e.key === 'Escape') {
      setOpen(false)
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn('w-full justify-between font-normal bg-input dark:bg-input/30', className)}
        >
          <div className='flex items-center'>
            <TagIcon className='mr-2 h-4 w-4' />
            {placeholder}
          </div>
          <ChevronsUpDown className="h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" align="start">
        <Command shouldFilter={false}>
          <CommandInput
            placeholder="Search tags..."
            value={searchValue}
            onValueChange={setSearchValue}
            onKeyDown={handleKeyDown}
          />
          <CommandList className="[&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-border [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:border-2 [&::-webkit-scrollbar-thumb]:border-transparent [&::-webkit-scrollbar-thumb]:bg-clip-padding hover:[&::-webkit-scrollbar-thumb]:bg-muted-foreground/30 [&::-webkit-scrollbar-thumb]:transition-colors">
            <CommandEmpty>
              {searchValue.trim() ? 'No tags found' : 'No tags available'}
            </CommandEmpty>
            {shouldShowCreateOption && (
              <CommandGroup>
                <CommandItem
                  onSelect={handleCreateTag}
                  disabled={isCreating}
                  className="cursor-pointer"
                >
                  {isCreating ? (
                    <>
                      <Spinner className="mr-2 h-4 w-4" />
                      Creating tag
                    </>
                  ) : (
                    <>
                      <Plus className="mr-2 h-4 w-4" />
                      Create tag: <span className="ml-1 font-medium">"{searchValue.trim()}"</span>
                    </>
                  )}
                </CommandItem>
              </CommandGroup>
            )}
            {shouldShowCreateOption && availableTagsForSelection.filter((tag) => {
              if (!searchValue.trim()) return true
              return tag.name.toLowerCase().includes(searchValue.toLowerCase())
            }).length > 0 && <CommandSeparator />}
            {availableTagsForSelection.filter((tag) => {
              if (!searchValue.trim()) return true
              return tag.name.toLowerCase().includes(searchValue.toLowerCase())
            }).length > 0 && (
              <CommandGroup>
              {availableTagsForSelection
                .filter((tag) => {
                  if (!searchValue.trim()) return true
                  return tag.name.toLowerCase().includes(searchValue.toLowerCase())
                })
                .map((tag) => (
                  <CommandItem
                    key={tag.id}
                    value={tag.name}
                    onSelect={() => handleTagSelect(tag)}
                    className="cursor-pointer"
                  >
                    <Check className={cn('mr-2 h-4 w-4', 'opacity-0')} />
                    {tag.name}
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}

