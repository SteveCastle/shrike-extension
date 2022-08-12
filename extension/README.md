**Getting Started**

Run the following commands to install dependencies and start developing

```
yarn install
yarn dev
```

**Scripts**

-   `yarn dev` - run `webpack` in `watch` mode
-   `yarn storybook` - runs the Storybook server
-   `yarn build` - builds the production-ready unpacked extension
-   `yarn test -u` - runs Jest + updates test snapshots
-   `yarn lint` - runs EsLint
-   `yarn prettify` - runs Prettier

<details>
  <summary>Loading the extension in Google Chrome</summary>

In [Google Chrome](https://www.google.com/chrome/), open up [chrome://extensions](chrome://extensions) in a new tab. Make sure the `Developer Mode` checkbox in the upper-right corner is turned on. Click `Load unpacked` and select the `dist` directory in this repository - your extension should now be loaded.

![Installed Extension in Google Chrome](https://i.imgur.com/Y2dQFte.png "Installed Extension in Google Chrome")

</details>

<details>
  <summary>Loading the extension in Brave</summary>

In [Brave](https://brave.com/), open up [brave://extensions](brave://extensions) in a new tab. Make sure the `Developer Mode` checkbox in the upper-right corner is turned on. Click `Load unpacked` and select the `dist` directory in this repository - your extension should now be loaded.

![Installed Extension in Brave](https://i.imgur.com/rKsbtcO.png "Installed Extension in Brave")

</details>

<details>
  <summary>Loading the extension in Mozilla Firefox</summary>

In [Mozilla Firefox](https://www.mozilla.org/en-US/firefox/new/), open up the [about:debugging](about:debugging) page in a new tab. Click the `This Firefox` link in the sidebar. One the `This Firefox` page, click the `Load Temporary Add-on...` button and select the `manfiest.json` from the `dist` directory in this repository - your extension should now be loaded.

![Installed Extension in Mozilla Firefox](https://i.imgur.com/FKfTw4B.png "Installed Extension in Mozilla Firefox")

</details>

<details>
  <summary>Loading the extension in Microsoft Edge</summary>

In [Microsoft Edge](https://www.microsoft.com/en-us/edge), open up [edge://extensions](edge://extensions) in a new tab. Make sure the `Developer Mode` checkbox in the lower-left corner is turned on. Click `Load unpacked` and select the `dist` directory in this repository - your extension should now be loaded.

![Installed Extension in Microsoft Edge](https://i.imgur.com/ykesx0g.png "Installed Extension in Microsoft Edge")

</details>
