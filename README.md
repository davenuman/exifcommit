
# EXIF Commit

A command-line tool to update the `ImageDescription` EXIF tag in image files using a commit-message style workflow.

## Description

`exifcommit` simplifies the process of batch updating the `ImageDescription` EXIF tag in image files. It provides a commit-message-like interface, allowing you to specify a new description in your preferred text editor and apply it to multiple image files at once.

**Key Features:**

*   **Batch Tagging:** Apply the same `ImageDescription` to multiple image files.
*   **Editor-Based Workflow:** Uses your favorite text editor to write the new description, offering flexibility and familiar editing features.
*   **File Selection by Search Term:**  Finds target image files based on a search term in their filenames.
*   **Preview of Current Tags:** Displays the current `ImageDescription` for each found file in the commit file, allowing you to verify targets before applying changes.
*   **Dry-Run (Implicit):**  By simply not saving the commit file in the editor, you can effectively perform a dry-run, reviewing the files that would be modified without making any changes.
*   **Uses `exiftool`:** Leverages the powerful `exiftool` command-line utility for efficient EXIF metadata manipulation.

## Installation

### Prerequisites

*   **Go:**  You need Go installed on your system. You can download it from [https://go.dev/dl/](https://go.dev/dl/). Ensure your Go environment is properly set up, including `$GOPATH` or `$GOBIN` being in your `$PATH`.
*   **exiftool:** `exifcommit` relies on the `exiftool` command-line utility. You need to install `exiftool` separately as it's not included with this Go program.  Instructions for installing `exiftool` can be found on the official website: [https://exiftool.org/install.html](https://exiftool.org/install.html). Make sure `exiftool` is in your system's `$PATH` so that `exifcommit` can execute it.

### Installing `exifcommit`

1.  **Clone the repository (if you have the source code):**
    ```bash
    git clone <repository_url>
    cd exifcommit
    ```

2.  **Install using `go install`:**
    ```bash
    go install github.com/davenuman/exifcommit/exifcommit@latest
    ```
    This command will download and build `exifcommit`, placing the executable in your `$GOPATH/bin` or `$GOBIN` directory. Ensure this directory is in your system's `$PATH` so you can run `exifcommit` from anywhere in the terminal.

## Usage

### Basic Usage

```bash
exifcommit <search_term>
```

Replace `<search_term>` with a term to search for in filenames. `exifcommit` will search for files in the current directory and any directories specified in the `EXIFCOMMIT_PATH` environment variable (see below). Filenames containing the `<search_term>` will be considered for EXIF tag modification.

**Example:**

To update the `ImageDescription` tag for all `.jpg` files in the current directory and its subdirectories, you could use:

```bash
exifcommit jpg
```

### Workflow

1.  **File Search:** `exifcommit` searches for files matching the `<search_term>` in the current directory and any paths defined by the `EXIFCOMMIT_PATH` environment variable.
2.  **Temporary Commit File:**  A temporary file (starting with `.EXIF-` and located in the current directory) is created. This file is opened in your default text editor (or the editor specified by the `EDITOR` environment variable).
3.  **Edit the Commit File:**
    *   **First Line:** The **first line** of the file is designated as the **new `ImageDescription`** that will be applied to the selected image files.
    *   **Empty Description:** If you leave the first line empty or delete its content, `exifcommit` will abort and no changes will be made.
    *   **File List:** The file contains a list of the files found, each preceded by `# file: <filepath>`. Below each file path, the current `ImageDescription` tag (if it exists) is displayed. Lines starting with `#` are treated as comments and are ignored during parsing, except for the `# file:` lines which identify the target file paths.
    *   **Modify Files:** To apply the new `ImageDescription` to a file, ensure the line starting with `# file: ...` for that file remains in the file and is not commented out.
    *   **Exclude Files:** To exclude a file from being modified, you can either **delete the entire line** starting with `# file: ...` or simply **comment it out** by adding another `#` at the beginning, making it `## file: ...`.
4.  **Save and Close Editor:** Save the changes in your text editor and close the editor.
5.  **Apply Changes:** `exifcommit` parses the temporary file. If a new description is provided and file lines are present, it uses `exiftool` to update the `ImageDescription` tag for the listed files using the provided description.
6.  **Cleanup:** The temporary commit file is automatically deleted after processing.

### Environment Variables

*   **`EXIFCOMMIT_PATH`:**  This environment variable allows you to extend the directories `exifcommit` searches in. It uses the system's path list separator (e.g., `:` on Linux/macOS, `;` on Windows).  You can specify multiple directories to search in by separating them with the path list separator.

    **Example:**

    ```bash
    export EXIFCOMMIT_PATH="$HOME/Pictures:$HOME/Documents/Images" # Linux/macOS
    set EXIFCOMMIT_PATH="$HOME/Pictures;$HOME/Documents\Images" # Windows
    ```

*   **`EDITOR`:**  This environment variable specifies the text editor `exifcommit` will use to open the commit file. If `EDITOR` is not set, `exifcommit` will default to `vim`. You can set it to your preferred editor (e.g., `nano`, `emacs`, `code`, `notepad++`).

    **Example:**

    ```bash
    export EDITOR="nano" # Use nano as the editor
    export EDITOR="code --wait" # Use VS Code (ensure `code` is in your PATH and `--wait` is used for proper workflow)
    ```

### Example Scenario

Let's say you want to add the description "Photos from my vacation in Hawaii" to all `.jpg` files in your `~/Pictures/Hawaii` directory.

1.  **Set `EXIFCOMMIT_PATH` (optional but helpful for organized searches):**

    ```bash
    export EXIFCOMMIT_PATH="$HOME/Pictures"
    ```

2.  **Run `exifcommit`:**

    ```bash
    exifcommit Hawaii.jpg
    ```

    This will:
    *   Search for files with "Hawaii.jpg" in their name within `~/Pictures` and the current directory.
    *   Create a temporary file and open it in your editor. The file might look something like this (contents are illustrative):

        ```
        # First line of this file is used for the ImageDescription
        # An empty line aborts the change.
        #
        # Files to be modified, and their current value
        # (remove to exclude from editing):
        # file: /home/user/Pictures/Hawaii/beach1.jpg
        # current description of beach1.jpg
        # file: /home/user/Pictures/Hawaii/beach2.jpg
        # another description
        # file: /home/user/Pictures/Hawaii/sunset.jpg
        # not found
        ```

3.  **Edit the Commit File:**
    *   On the **first line**, type in the new description: `Photos from my vacation in Hawaii`
    *   Review the file list. If you want to exclude `beach2.jpg` from being modified, you can either delete the line `# file: /home/user/Pictures/Hawaii/beach2.jpg` or comment it out like `## file: /home/user/Pictures/Hawaii/beach2.jpg`.
    *   The file should now look like this (if you kept all files and added the description):

        ```
        Photos from my vacation in Hawaii
        # First line of this file is used for the ImageDescription
        # An empty line aborts the change.
        #
        # Files to be modified, and their current value
        # (remove to exclude from editing):
        # file: /home/user/Pictures/Hawaii/beach1.jpg
        # current description of beach1.jpg
        # file: /home/user/Pictures/Hawaii/beach2.jpg
        # another description
        # file: /home/user/Pictures/Hawaii/sunset.jpg
        # not found
        ```

4.  **Save and Close:** Save the file and close your editor.
5.  **Verification:** `exifcommit` will now apply the "Photos from my vacation in Hawaii" description to `beach1.jpg` and `sunset.jpg` (and `beach2.jpg` if you didn't exclude it). You can verify the changes using an EXIF viewer or by running `exifcommit` again to check the current descriptions.

**Important Notes:**

*   **Backup Files:** `exiftool` might create backup files of your original images (typically with a `.original` extension) in the same directory. `exifcommit` attempts to delete these backup files after successfully writing the EXIF tags using the `-delete_original!` flag of `exiftool`.
*   **Error Handling:** `exifcommit` provides basic error handling and will print warnings or error messages to the console if issues occur. Check the output for any potential problems.
*   **`exiftool` Compatibility:**  Ensure that `exiftool` is functioning correctly on your system independently before using `exifcommit`.

This README provides a comprehensive guide to installing and using `exifcommit`. Enjoy tagging your images!
