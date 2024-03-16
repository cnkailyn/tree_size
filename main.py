import os
import time


class Colors:
    GREEN = '\033[92m'
    WHITE = '\033[97m'
    RESET = '\033[0m'

size_cache = {}

def format_size(size_bytes):
    if size_bytes < 1024:
        return f"{size_bytes} B"
    elif size_bytes < 1024 ** 2:
        return f"{size_bytes / 1024:.2f} KB"
    elif size_bytes < 1024 ** 3:
        return f"{size_bytes / 1024 ** 2:.2f} MB"
    elif size_bytes < 1024 ** 4:
        return f"{size_bytes / 1024 ** 3:.2f} GB"
    else:
        return f"{size_bytes / 1024 ** 4:.2f} TB"

def get_size(start_path = '.'):
    if start_path in size_cache:
        return size_cache[start_path]

    total_size = 0
    for dirpath, _, filenames in os.walk(start_path):
        for f in filenames:
            fp = os.path.join(dirpath, f)
            if not os.path.islink(fp):
                total_size += os.path.getsize(fp)
    size_cache[start_path] = total_size
    return total_size

def parse_size(size_str):
    units = {'B': 1, 'KB': 1024, 'MB': 1024**2, 'GB': 1024**3, 'TB': 1024**4}
    size_str = size_str.upper().replace(' ', '')
    number, unit = float(size_str[:-2]), size_str[-2:]
    return int(number * units[unit])

def print_tree(start_path, prefix='', depth=-1, file_include=None, file_exclude=None, min_size=None, max_size=None, current_depth=0, stats=None, only_path=False):
    start_time = time.time()
    if stats is None:
        stats = {'folders': 0, 'files': 0}

    if file_include is None:
        file_include = []
    if file_exclude is None:
        file_exclude = []

    if depth >= 0 and current_depth > depth:
        return

    size_of_current_path = get_size(start_path)
    formatted_size = format_size(size_of_current_path)
    if prefix == '':
        print(f"{Colors.GREEN}{start_path} ({formatted_size}){Colors.RESET}")
    else:
        print(f"{prefix}├── {Colors.GREEN}{os.path.basename(start_path)} ({formatted_size}){Colors.RESET}")

    prefix = prefix.replace("├──", "│  ").replace("└──", "   ")
    file_list = os.listdir(start_path)
    for i, file in enumerate(sorted(file_list)):
        path = os.path.join(start_path, file)
        if os.path.isdir(path):
            stats['folders'] += 1
            size_in_mb = get_size(path)
            formatted_size = format_size(size_in_mb)
            if (min_size is None or size_in_mb >= parse_size(min_size)) and (max_size is None or size_in_mb <= parse_size(max_size)):
                if i == len(file_list) - 1:
                    print_tree(path, prefix + "└── ", depth, file_include, file_exclude, min_size, max_size, current_depth + 1, stats, only_path)
                else:
                    print_tree(path, prefix + "├── ", depth, file_include, file_exclude, min_size, max_size, current_depth + 1, stats, only_path)
        elif not only_path:
            stats['files'] += 1
            file_extension = os.path.splitext(file)[1][1:].lower()
            if file_extension in map(str.lower, file_exclude) or (file_include and file_extension not in map(str.lower, file_include)):
                continue
            
            size_in_bytes = os.path.getsize(path)
            formatted_size = format_size(size_in_bytes)
            if (min_size is None or size_in_bytes >= parse_size(min_size)) and (max_size is None or size_in_bytes <= parse_size(max_size)):
                if i == len(file_list) - 1:
                    print(f"{prefix}└── {Colors.WHITE}{file} ({formatted_size}){Colors.RESET}")
                else:
                    print(f"{prefix}├── {Colors.WHITE}{file} ({formatted_size}){Colors.RESET}")

    if current_depth == 0:
        print()
        print("Summarize:")
        print(f"Scaned paths: {stats['folders']}")
        print(f"Scaned files: {stats['files']}")
        end_time = time.time()
        print(f"Cost time: {end_time - start_time:.2f} s")
        

if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description='Print the directory tree')
    parser.add_argument('path', nargs='?', default='.', help='The starting path, defaults to the current directory')
    parser.add_argument('--depth', type=int, default=999, help='Depth of the directory tree, defaults is 999')
    parser.add_argument('--file-include', nargs='*', default=[], help='List of file types to include')
    parser.add_argument('--file-exclude', nargs='*', default=[], help='List of file types to exclude')
    parser.add_argument('--min-size', default=None, help='Minimum file size, defaults to none (unlimited), example: 10KB, 20MB')
    parser.add_argument('--max-size', default=None, help='Maximum file size, defaults to none (unlimited), example: 10KB, 20MB')
    parser.add_argument('--only-path', action='store_true', help='Whether to only print paths, defaults to False')

    args = parser.parse_args()

    print_tree(start_path=args.path, 
               depth=args.depth, 
               file_include=args.file_include, 
               file_exclude=args.file_exclude, 
               min_size=args.min_size, 
               max_size=args.max_size, 
               only_path=args.only_path)
