DP is Density-independent pixels: an abstract unit that is based on the physical density of the screen.
These units are relative to a 160 dpi (dots per inch) screen, on which 1 dp is roughly equal to 1 px.
When running on a higher density screen, the number of pixels used to draw 1 dp is scaled up by a factor
appropriate for the screen's dpi.

Likewise, when on a lower-density screen, the number of pixels used for 1 dp is scaled down.
The ratio of dps to pixels changes with the screen density, but not necessarily in direct proportion.
Using dp units instead of px units is a solution to making the view dimensions in your layout
resize properly for different screen densities. It provides consistency for the real-world sizes of
your UI elements across different devices.
Source: https://developer.android.com/guide/topics/resources/more-resources.html#Dimension