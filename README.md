## Wordepress

Wordepress is a tool for injecting collections of markdown formatted
documentation (typically technical documentation stored in git repositories)
into a WordPress installation so that it can be integrated seamlessly with an
organisation's existing web presence.

Features:

* Keep documentation content with the code
* Manage presentation in WordPress
* Leverage SEO & tracking features of WordPress

## Installation

### Install CLI

You must have a working Go installation - see the
[Go Documentation](https://golang.org/doc/install) for details. Make
sure you can install binaries with `go install`, setting your `$PATH`
and `$GOPATH`/`$GOBIN` as necessary. Then:

    $ go get github.com/weaveworks/wordepress/cmd/wordepress

You should now be able to execute `wordepress`:

```
$ wordepress
Technical documentation importer for WordPress

Usage:
  wordepress [command]

Available Commands:
  publish     Publish a site into WordPress
  delete      Delete a site from WordPress

Flags:
  -h, --help   help for wordepress

Use "wordepress [command] --help" for more information about a command.
```

### Configure WordPress

This configuration needs to be performed once only for a given
WordPress installation, before the first set of documents is uploaded.

#### Activate Plugins

The following third-party plugins must be installed and activated:

* Toolset Types, Views & Layouts
* WordPress REST API (Version 2)
* Application Passwords

Then the Wordepress plugin must be installed and activated. The
following commands will produce a ZIP file which can be uploaded via
the WordPress admin UI:

* `cd wordepress/plugin`
* `zip -r wordepress.zip .`

#### Import Types, Layouts & Views

First you must import the Wordepress custom post type, views and
layout via the WordPress admin UI. These components rely on the Toolset
commercial Types, Layouts & Views plugins being installed and
activated.

> NB the layout depends on the views and types, so be sure to import
> them in the mandated order:

* Navigate to Types -> Dashboard -> Import/Export -> Import File and import
  [wordepress/toolset-exports/wordpress-types.zip](toolset-exports/wordepress-types.zip)
* Navigate to Views -> Import/Export -> Import and import
  [wordepress/toolset-exports/wordepress-views.zip](toolset-exports/wordepress-views.zip)
* Navigate to Layouts -> Import/Export -> Import and import
  [wordepress/toolset-exports/wordepress-layouts.zip](toolset-exports/wordepress-layouts.zip)
  Make sure to check 'Overwrite layout assignments' so that the layout
  becomes the default template for all documentation pages.
* Navigate to Views -> Settings -> Compatability and add
  `wp_get_post_parent_id` to the list of functions which can be used
  inside conditional evaluations

Finally, if your site uses a top level layout (to provide standard
headers and footers for example) you may wish to configure it to be
the parent of the Wordepress documentation layout.

## Usage

First you must add an application password to your WordPress account
that the CLI tool will use to authenticate itself with the REST API -
this is so you can avoid exposing your WordPress account password:

* Navigate to Users -> Your Profile -> Application Passwords
* Enter 'wordepress' as the application name and choose 'Add New'

You can now use the `wordepress` command to inject documentation into
your WordPress instance.

> NB the `--user` parameter is the WordPress username under which you
> configured the application password (not the application password
> name) whilst the `--password` parameter is the value emitted by 'Add
> New' above

    wordepress publish --url https://dev-weavewww.pantheon.io \
        --user <wordpress-username> \
        --password <generated-application-password> \
        --product net --version 1.5 \
        ~/workspace/weave/site

## Repo Format

* Each page is a markdown file ending in `.md`
* The filename sans extension is used as the URL path segment (the
  Wordpress 'slug') in the uploaded document
* Each `.md` file underneath `/site`is uploaded as a toplevel page.
  For each such markdown file, if there exists a subdirectory with the
  same name (sans the `.md` extension) its contents are processed
  recursively and all pages within it are uploaded as children of the
  toplevel page. Beware that the Wordpress view that generates the
  navigation only reflects the top two levels however, so don't nest
  pages deeper than one level from the top
* Any intrasite links must be relative URLS (e.g. path only) whose
  path is absolute with respect to the git repository root e.g.
  `/site/foo.md`. Relative paths (e.g. foo.md, or ../foo.md) are not
  yet supported (see #8)
* Conversely, at the moment images must live in the same directory as
  the markdown file they're beeing linked from and must be referred to
  with an unqualified filename e.g. `foo.png` (see #8)
* Images must currently be in PNG format (see #20)

Finally, each markdown file requires a header block:

```
---
title: My Document Title
menu_order: 10
---
```

to control the Wordpress page title and the order in which pages
appear in the navigation.
