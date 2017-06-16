Can compute the discrete fourier transform of an image as an image, and can apply the inverse DFT to the resultant image and get the original image back. Currently only works on images with power-of-two dimensions.

Usage:

Generate spectrum of image (resultant spectrum only encodes brightness of pixels, so performing the inverse DFT will result in a black and white image):
`imagedft <file>`

Generate original signal from spectrum image:
`imagedft -i <file>`
